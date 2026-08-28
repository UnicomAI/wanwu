// Package skill_var 提供 skill 自定义变量的校验规则与共享常量，
// 供 BFF 入口与 wga-sandbox 兜底两侧共用，
// 单一事实源 —— 黑名单更新时不会两边漂移。
package skill_var

import (
	"fmt"
	"regexp"
	"strings"
)

// MaxVariableValueLen 单条 variableValue 的最大字节长度，覆盖典型 API key/token/短 JSON。
const MaxVariableValueLen = 16 * 1024

var (
	// VariableKeyPattern 合法环境变量名 (POSIX env name)。
	VariableKeyPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

	// VariableKeyDeny BFF Check 与 sandbox collectSkillEnvVars 共用的黑名单，单一事实源。
	// 收录条件：可被利用做代码执行 / 影响 dynamic linker / 影响 shell 启动行为。
	// 仅身份/路径类变量 (USER / HOSTNAME / PWD / ...) 不在此列 —— 它们不是可执行路径，
	// 对允许任意 shell 的 sandbox 没有边际收益，拦了反而干扰用户的合法配置。
	VariableKeyDeny = map[string]struct{}{
		"PATH":            {},
		"LD_PRELOAD":      {},
		"LD_LIBRARY_PATH": {},
		"SHELL":           {},
		"HOME":            {},
		"IFS":             {},
		"PS1":             {},
		"BASH_ENV":        {}, // bash 启动时自动 source 指定文件 — 等同代码注入
		"ENV":             {}, // POSIX shell 对 BASH_ENV 的等价物
		"PROMPT_COMMAND":  {}, // 每次显示 prompt 前自动 exec
		"LD_AUDIT":        {}, // glibc 在动态链接阶段加载共享库 (绕过 LD_PRELOAD 黑名单)
		"LD_DEBUG":        {}, // 触发 dl 调试输出，可能泄露文件路径
	}
)

// ValidateVariableKey 校验 key 是否符合 env 命名规则且不在黑名单。
// 错误信息包含 key 名但绝不包含 value。
func ValidateVariableKey(key string) error {
	if !VariableKeyPattern.MatchString(key) {
		return fmt.Errorf("variableKey %q invalid: must match %s", key, VariableKeyPattern)
	}
	if _, deny := VariableKeyDeny[strings.ToUpper(key)]; deny {
		return fmt.Errorf("variableKey %q is reserved", key)
	}
	return nil
}

// ValidateVariableValueLen 校验 value 长度上限。不打印 value 内容。
func ValidateVariableValueLen(val string) error {
	if len(val) > MaxVariableValueLen {
		return fmt.Errorf("variableValue exceeds max length %d", MaxVariableValueLen)
	}
	return nil
}
