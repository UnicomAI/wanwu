package shared

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/cloudwego/eino/adk/filesystem"
)

const (
	maxOutputSize  = 1024 * 1024 // 1MB
	commandTimeout = 5 * time.Minute
)

// SkillEnvFileName sandbox workspace 下注入 skill 变量的文件名。
// 写侧（eino runner.injectEnvVariables）与读侧（ShellOnlyBackend.Execute）共用此常量，
// 避免两处字面值漂移。JSON 格式不与 dotenv 兼容，故文件名带 .json 后缀。
const SkillEnvFileName = ".skill_env.json"

// ShellOnlyBackend 仅提供 shell 命令执行能力。
// 不实现 filesystem.Backend 的其他文件操作方法，因为 bash 工具只调用 Execute。
type ShellOnlyBackend struct {
	maxOutputSize  int
	commandTimeout time.Duration
	workDir        string
	// skillEnvPath 指向 workspace 下的 SkillEnvFileName 文件；Execute 每次会重新读，
	// 文件不存在时保持 cmd.Env=nil（透明继承父进程 env），从而对无 skill var 的请求零侵入。
	skillEnvPath string
	// halt 累计单次 sandbox 会话内的连续 [BLOCKED:...] 次数；连续到阈值触发熔断。
	// 可空：nil 表示不启用熔断（兼容 tests + oneshot 沙箱路径）。
	halt *HaltState
}

// NewShellOnlyBackend 构造 ShellOnlyBackend。halt 可空——为 nil 时不启用连续 BLOCKED 熔断。
func NewShellOnlyBackend(workDir string, halt *HaltState) *ShellOnlyBackend {
	return &ShellOnlyBackend{
		maxOutputSize:  maxOutputSize,
		commandTimeout: commandTimeout,
		workDir:        workDir,
		skillEnvPath:   filepath.Join(workDir, SkillEnvFileName),
		halt:           halt,
	}
}

// haltedResponse 返回熔断后 Execute 的短路响应。命令不进入 cmd.Run，直接向 LLM 返回终止提示。
func haltedResponse() *filesystem.ExecuteResponse {
	exitCode := 1
	return &filesystem.ExecuteResponse{
		Output:   "[BLOCKED:HALT] session terminated by sandbox security guard due to repeated policy violations; no further commands will execute.",
		ExitCode: &exitCode,
	}
}

// readSkillEnv 读取并反序列化 .skill_env.json。文件不存在返回 nil（合法路径）；
// 存在但解析失败打日志后返回 nil（不阻断命令执行，保持"无 var = 透明继承"的兜底行为）。
// 错误信息只带文件路径，不含内容片段。
func (b *ShellOnlyBackend) readSkillEnv() map[string]string {
	if b.skillEnvPath == "" {
		return nil
	}
	data, err := os.ReadFile(b.skillEnvPath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("[Shell] read %s failed: %v", b.skillEnvPath, err)
		}
		return nil
	}
	if len(data) == 0 {
		return nil
	}
	var envMap map[string]string
	if err := json.Unmarshal(data, &envMap); err != nil {
		log.Printf("[Shell] parse %s failed: %v", b.skillEnvPath, err)
		return nil
	}
	return envMap
}

// Execute 执行 shell 命令，附带安全校验、超时、输出截断与退出码处理。
//
// 三层防线（按顺序）：
//  1. halt 短路：已熔断则直接返回 [BLOCKED:HALT]，命令永不进入 cmd.Run。
//  2. validateCommand：通用系统安全（rm -rf / 路径穿越 / 写敏感文件等）。
//     与 skill vars 状态无关，**永远生效**——predates 本次 skill-var 工作。
//  3. precheck 家族（precheckCommand + precheckScriptFile）：skill 变量保护层。
//     仅在 skill 配置了至少一个 var（envMap 非空）时启用——**无 vars 时跳过**，
//     backward compat 到 v4 之前的行为。
//
// 熔断计数：precheck 拦截或 validate 拦截触发的 output 会以 [BLOCKED:xxx] 或
// `安全拦截：` 起头；连续 BLOCKED 到阈值触发 haltFn（绑定到 sessionCtx 的
// CancelCauseFunc），eino runner iterator 关闭，上层通过 BuildFinalAgentEvent
// 下发 error[agent] 兜底 SSE。**halt 计数同样只在有 skill vars 时启用**——
// 无 vars 时 precheck 不跑，halt 也无意义；validate 拦截在无 vars 场景不计入
// halt（保持 backward compat 语义纯粹）。
func (b *ShellOnlyBackend) Execute(ctx context.Context, req *filesystem.ExecuteRequest) (*filesystem.ExecuteResponse, error) {
	// 已熔断：短路返回，命令不进入执行链。放在最前面确保后续步骤都不跑。
	if b.halt != nil && b.halt.Halted() {
		return haltedResponse(), nil
	}

	// 通用安全校验（与 skill vars 状态无关；可由 DISABLE_EINO_GENERIC_GUARD=1 关闭）
	if !genericGuardDisabled() {
		if err := validateCommand(req.Command); err != nil {
			exitCode := 1
			return &filesystem.ExecuteResponse{
				Output:   err.Error(),
				ExitCode: &exitCode,
			}, nil
		}
	}

	// 读取 skill 已注入的环境变量。文件格式：JSON map[string]string（由 runner BeforeRun
	// 的 injectEnvVariables 写入）。两条不变量：
	//   1. 文件不存在时 cmd.Env 保持 nil（Go exec 自动继承父进程 env），透明无侵入；
	//   2. 文件存在时 cmd.Env = append(os.Environ(), kv...)，PATH/PYTHONPATH/HOME 不丢失，
	//      同名 key 以 skill var 为准（os/exec 取后出现的值）。
	// 安全：读 / 解析失败的错误信息只带文件路径不带值；同样不打印 envMap 内容。
	envMap := b.readSkillEnv()
	hasVars := len(envMap) > 0

	// Skill 变量保护层：仅在有 skill vars 时启用。
	// 无 vars 时保护对象不存在，precheck 跳过——backward compat 到 v4 之前的行为。
	// 可由 DISABLE_EINO_SKILL_VAR_GUARD=1 关闭（调试 / 特殊场景）。
	if hasVars && !skillVarGuardDisabled() {
		// cmd.Run 前对命令字符串做静态拦截，阻止 LLM 通过 echo/cat/printenv 等方式回流敏感 value。
		if err := precheckCommand(req.Command, envMap); err != nil {
			exitCode := 1
			resp := &filesystem.ExecuteResponse{
				Output:   err.Error(),
				ExitCode: &exitCode,
			}
			b.recordHalt(resp.Output, hasVars)
			return resp, nil
		}
		// 文件形式脚本 body-scan：拦下 `python3 x.py` / `bash script.sh` 等通过文件运行脚本的越权手法。
		if err := precheckScriptFile(req.Command, b.workDir, envMap); err != nil {
			exitCode := 1
			resp := &filesystem.ExecuteResponse{
				Output:   err.Error(),
				ExitCode: &exitCode,
			}
			b.recordHalt(resp.Output, hasVars)
			return resp, nil
		}
	}

	execCtx, cancel := context.WithTimeout(ctx, b.commandTimeout)
	defer cancel()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(execCtx, "cmd", "/C", req.Command)
	} else {
		cmd = exec.CommandContext(execCtx, "sh", "-c", req.Command)
	}

	if b.workDir != "" {
		if absWorkDir, err := filepath.Abs(b.workDir); err == nil {
			cmd.Dir = absWorkDir
		}
	}

	if len(envMap) > 0 {
		env := os.Environ()
		for k, v := range envMap {
			env = append(env, k+"="+v)
		}
		cmd.Env = env
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	if execCtx.Err() == context.DeadlineExceeded {
		exitCode := 124
		return &filesystem.ExecuteResponse{
			Output:   fmt.Sprintf("命令执行超时（限制 %v），已终止。请考虑拆分任务或优化命令。", b.commandTimeout),
			ExitCode: &exitCode,
		}, nil
	}

	output := stdout.String()
	if stderr.Len() > 0 {
		if len(output) > 0 {
			output += "\n"
		}
		output += stderr.String()
	}

	truncated := false
	if len(output) > b.maxOutputSize {
		output = output[:b.maxOutputSize]
		truncated = true
	}

	exitCode := 0
	if err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if !ok {
			return nil, err
		}
		exitCode = exitErr.ExitCode()
	}

	resp := &filesystem.ExecuteResponse{
		Output:    output,
		ExitCode:  &exitCode,
		Truncated: truncated,
	}
	// exec 正常收尾也走 recordHalt：output 若是 sh 报错回显"[BLOCKED:..."/"安全拦截："
	// 之类的字符串（合法场景下几乎不可能，兜底），也一并计入；否则 NoteSuccess 重置计数。
	b.recordHalt(resp.Output, hasVars)
	return resp, nil
}

// recordHalt 根据 output 是否为拦截产物累计或重置 halt 计数。
// hasVars=false 时短路——无 skill vars 场景下 halt 不启用（backward compat）。
func (b *ShellOnlyBackend) recordHalt(output string, hasVars bool) {
	if b.halt == nil || !hasVars {
		return
	}
	if isBlockedOutput(output) {
		b.halt.NoteBlocked()
	} else {
		b.halt.NoteSuccess()
	}
}

// isBlockedOutput 判断 Execute 的 Output 是否为安全拦截产物。
// 覆盖两个前缀：precheck 家族的 [BLOCKED:xxx] 与 validateCommand 家族的 `安全拦截：...`。
func isBlockedOutput(output string) bool {
	return strings.HasPrefix(output, "[BLOCKED:") || strings.HasPrefix(output, "安全拦截：")
}
