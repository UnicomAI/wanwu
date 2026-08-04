package eino

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/UnicomAI/wanwu/pkg/log"
	skill_var "github.com/UnicomAI/wanwu/pkg/skill-var"
	"github.com/UnicomAI/wanwu/pkg/wga-sandbox/internal/runner/eino/agent/shared"
	wga_sandbox_option "github.com/UnicomAI/wanwu/pkg/wga-sandbox/wga-sandbox-option"
)

// setupEnv 写入沙箱内 .env 文件，供 eino-agent 与 bash 进程读取模型 / trace 配置。
func (r *Runner) setupEnv(ctx context.Context) error {
	var lines []string

	if r.req.ModelConfig.APIKey != "" {
		lines = append(lines, fmt.Sprintf("OPENAI_API_KEY=%s", r.req.ModelConfig.APIKey))
	}

	// 提取一次 traceparent，复用给 baseURL 拼接与 TRACEPARENT env 行。
	traceParent := ""
	if r.req.TraceContext != nil {
		traceParent = r.req.TraceContext["traceparent"]
	}

	baseURL := r.req.ModelConfig.BaseURL
	// 若存在 traceparent，把 traceId/spanId 编码到 baseURL 路径里。
	// BFF 侧新增带 /trace/:traceId/span/:spanId/ 参数的路由来接收这种请求。
	if traceParent != "" {
		if parts := strings.Split(traceParent, "-"); len(parts) == 4 {
			baseURL = baseURL + "/trace/" + parts[1] + "/span/" + parts[2]
		}
	}
	if baseURL != "" {
		lines = append(lines, fmt.Sprintf("OPENAI_BASE_URL=%s", baseURL))
	}
	if r.req.ModelConfig.Model != "" {
		lines = append(lines, fmt.Sprintf("OPENAI_MODEL_ID=%s", r.req.ModelConfig.Model))
	}

	// LLM 采样参数（temperature/topP/maxTokens 等）以 JSON 序列化成单行写入。
	// 沙箱内 resolveAgent 读回后 json.Unmarshal 进 AppConfig.Params。
	// 用 JSON 而非多行 env：规避 godotenv 对含 $ 值的变量展开，且 *Enable 开关一并携带。
	if r.req.ModelConfig.Params != nil {
		if data, err := json.Marshal(r.req.ModelConfig.Params); err == nil {
			lines = append(lines, fmt.Sprintf("OPENAI_LLM_PARAMS=%s", string(data)))
		} else {
			log.Warnf("%s marshal ModelConfig.Params failed: %v", r.logPrefix, err)
		}
	}

	// 追加 trace 环境变量，供沙箱内 bash 进程（含 curl 调用）继续传播 trace。
	if traceParent != "" {
		lines = append(lines, fmt.Sprintf("TRACEPARENT=%s", traceParent))
	}
	if r.req.TraceContext != nil {
		if ts := r.req.TraceContext["tracestate"]; ts != "" {
			lines = append(lines, fmt.Sprintf("TRACESTATE=%s", ts))
		}
		if bg := r.req.TraceContext["baggage"]; bg != "" {
			lines = append(lines, fmt.Sprintf("BAGGAGE=%s", bg))
		}
	}

	content := strings.Join(lines, "\n") + "\n"
	if err := r.sb.WriteFile(ctx, ".env", []byte(content)); err != nil {
		return fmt.Errorf("failed to create .env: %w", err)
	}
	log.Infof("%s .env file created in sandbox workspace", r.logPrefix)
	return nil
}

// setupWorkspaceDirs 创建 skills/、input/、output/、tmp/ 目录，并把宿主 skills 与 input 复制进沙箱。
// eino-agent HTTP 服务从 workspace/skills/ 加载技能。
// input/ 无论是否有用户输入都必须先建好：agent 系统提示会声明三个目录已物理存在并禁止 LLM 自建，
// LLM 用 ls {{.Workspace}}/input/ 定位用户文件，目录缺失会导致定位失败。
func (r *Runner) setupWorkspaceDirs(ctx context.Context) error {
	if _, err := r.sb.ExecuteSync(ctx, "mkdir", "-p", "skills"); err != nil {
		return fmt.Errorf("failed to create skills directory: %w", err)
	}

	for _, skill := range r.req.Skills {
		log.Infof("%s copying skill from %s to skills/", r.logPrefix, skill.Dir)
		if err := r.sb.CopyToSandbox(ctx, skill.Dir, "skills"); err != nil {
			return fmt.Errorf("failed to copy skill to workspace: %w", err)
		}
	}

	// input/ 必须先建好，再把用户输入复制进去（落到 input/ 子目录而非 workspace 根）。
	if _, err := r.sb.ExecuteSync(ctx, "mkdir", "-p", "input"); err != nil {
		return fmt.Errorf("failed to create input directory: %w", err)
	}
	if r.req.InputDir != "" {
		log.Infof("%s copying input from %s to input/", r.logPrefix, r.req.InputDir)
		if err := r.sb.CopyToSandbox(ctx, r.req.InputDir, "input"); err != nil {
			return fmt.Errorf("failed to copy input to workspace: %w", err)
		}
	}

	if _, err := r.sb.ExecuteSync(ctx, "mkdir", "-p", "output"); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	if _, err := r.sb.ExecuteSync(ctx, "mkdir", "-p", "tmp"); err != nil {
		return fmt.Errorf("failed to create tmp directory: %w", err)
	}
	return nil
}

// --- Skill 变量注入 ---
//
// 隔离与安全约束：
//   - VariableValue 仅允许写入 sandbox workspace 下 .skill_env.json（供 ShellOnlyBackend 读回注入 cmd.Env），
//     绝不写入 SKILL.md / 日志 / 错误信息。
//   - VariableKey 校验走 skill_var.ValidateVariableKey 共享模块（与 BFF Check() 单一事实源）。
//   - 跨 skill VariableKey 撞键直接返回 error，BeforeRun 失败。当前调用方 buildSkillOptions
//     永远只塞 1 个 skill，撞键不会发生；这个断言是为了在未来真有人改 buildSkillOptions
//     塞多个 skill 时把问题拦在 prepare 阶段——倒逼那时再做 "<prefix>_<KEY>" 命名隔离方案。
//   - 被过滤掉的 key 同步不出现在 .skill_env.json 与 SKILL.md，避免 LLM 引用一个其实未注入的 key。
//   - 文件格式选用 JSON 而非 dotenv：dotenv parser (godotenv) 对双引号值无条件做 $VAR 展开，
//     会把含 $ 的密钥静默截断；JSON 无这层歧义，且支持任意字符的 round-trip。

// collectSkillEnvVars 把 r.req.Skills[*].Variables 扁平化为 KV map，应用校验与冲突策略。
// 返回 (sortedKeys, kv, err)，sortedKeys 保证写文件顺序稳定（便于测试与排查），且不打印 value。
// 跨 skill 同 VariableKey 时直接返回 error，让 BeforeRun 失败——理由见上方注释。
func (r *Runner) collectSkillEnvVars() ([]string, map[string]string, error) {
	kv := make(map[string]string)
	for _, skill := range r.req.Skills {
		for _, v := range skill.Variables {
			if err := skill_var.ValidateVariableKey(v.VariableKey); err != nil {
				log.Warnf("%s skip variable: %v", r.logPrefix, err)
				continue
			}
			if _, dup := kv[v.VariableKey]; dup {
				// 错误信息只带 key 名，不带 value，避免泄露。
				return nil, nil, fmt.Errorf("skill VariableKey collision across skills: %q "+
					"(current buildSkillOptions only packs 1 skill per sandbox run; "+
					"if you're packing multiple, prefix keys per skill to isolate)", v.VariableKey)
			}
			kv[v.VariableKey] = v.VariableValue
		}
	}
	keys := make([]string, 0, len(kv))
	for k := range kv {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys, kv, nil
}

// injectEnvVariables 把通过校验的 skill 变量以 JSON 格式写入 sandbox workspace 下的 .skill_env.json。
// ShellOnlyBackend 在每次 bash Execute 前 json.Unmarshal 这个文件并合入 cmd.Env。
func (r *Runner) injectEnvVariables(ctx context.Context) error {
	keys, kv, err := r.collectSkillEnvVars()
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		return nil
	}
	data, err := json.Marshal(kv)
	if err != nil {
		// 错误信息只带 key 数量，不带 key/value，避免泄露。
		return fmt.Errorf("failed to marshal %s (%d entries): %w", shared.SkillEnvFileName, len(keys), err)
	}
	if err := r.sb.WriteFile(ctx, shared.SkillEnvFileName, data); err != nil {
		return fmt.Errorf("failed to write %s (%d entries): %w", shared.SkillEnvFileName, len(keys), err)
	}
	log.Infof("%s %s written with %d entries (keys=%v)", r.logPrefix, shared.SkillEnvFileName, len(keys), keys)
	return nil
}

// setupSkillVariables 往每个 skill 的 SKILL.md 末尾追加 "## Variables" 段，
// 只展示 key / name / desc，**绝不展示 value**——避免明文密钥进入 LLM 上下文。
//
// 实现走 "cat → 拼 → WriteFile 覆盖" 模式而非 shell 追加，原因：
// reuseSandbox.ExecuteSync 的实现是 strings.Join(args, " ") 后整条扔给容器内 shell；
// 我们传 "sh", "-c", appendCmd 时，appendCmd 里的 pipe / 重定向 / 引号会被容器 shell
// 重新解析，sh -c 仅吃到 appendCmd 的第一个 token —— 历史 bug。
// WriteFile 走二进制 uploadData，不经 shell，所有引号/pipe 陷阱自动规避。
func (r *Runner) setupSkillVariables(ctx context.Context) error {
	for _, skill := range r.req.Skills {
		if len(skill.Variables) == 0 {
			continue
		}
		appendix := formatVariablesMarkdown(skill.Variables)
		if appendix == "" {
			continue // 所有 key 都被过滤
		}
		dirName := path.Base(skill.Dir)
		skillMDPath := fmt.Sprintf("skills/%s/SKILL.md", dirName)

		// ExecuteSync("cat", path) 经 strings.Join 后就是 "cat skills/<name>/SKILL.md"，
		// path 不含 pipe/引号，容器 shell 直接 exec cat，行为可控。
		existing, err := r.sb.ExecuteSync(ctx, "cat", skillMDPath)
		if err != nil {
			// SKILL.md 缺失就跳过本 skill，不阻断别的 skill 的 prepare。
			log.Warnf("%s skip SKILL.md variables append for %s: cat failed: %v",
				r.logPrefix, dirName, err)
			continue
		}
		final := existing + appendix

		if err := r.sb.WriteFile(ctx, skillMDPath, []byte(final)); err != nil {
			return fmt.Errorf("failed to write SKILL.md for %s: %w", dirName, err)
		}
		log.Infof("%s SKILL.md variables section appended for %s", r.logPrefix, dirName)
	}
	return nil
}

// formatVariablesMarkdown 生成 SKILL.md 追加段。
// 安全约束（不变量）：函数实现中不得引用 v.VariableValue；测试中 grep 任何 value 应得 0 命中。
func formatVariablesMarkdown(vars []wga_sandbox_option.SkillVariable) string {
	var b strings.Builder
	wrote := false
	for _, v := range vars {
		if err := skill_var.ValidateVariableKey(v.VariableKey); err != nil {
			continue
		}
		if !wrote {
			b.WriteString("\n## Variables\n\n")
			b.WriteString("The following variables are pre-injected into the bash environment as environment variables.\n")
			b.WriteString("In shell commands use `$KEY`; in Python scripts use `os.environ[\"KEY\"]`.\n")
			b.WriteString("Do NOT echo or print these values; they are sensitive.\n\n")
			wrote = true
		}
		name := v.Name
		if name == "" {
			name = v.VariableKey
		}
		if v.Description != "" {
			fmt.Fprintf(&b, "- **%s** — %s: `$%s`\n", name, v.Description, v.VariableKey)
		} else {
			fmt.Fprintf(&b, "- **%s**: `$%s`\n", name, v.VariableKey)
		}
	}
	if wrote {
		b.WriteString("\n")
	}
	return b.String()
}
