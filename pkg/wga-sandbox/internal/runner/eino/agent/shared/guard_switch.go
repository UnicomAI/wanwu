package shared

import "os"

// genericGuardDisabled 读取 DISABLE_EINO_GENERIC_GUARD。
// 返回 true 表示关闭通用安全拦截（validateCommand）。
// 默认未设置 = 拦截生效；显式置 "1"/"true" 才关闭。
func genericGuardDisabled() bool {
	v := os.Getenv("DISABLE_EINO_GENERIC_GUARD")
	return v == "1" || v == "true"
}

// skillVarGuardDisabled 读取 DISABLE_EINO_SKILL_VAR_GUARD。
// 返回 true 表示关闭 Skill 变量保护拦截（precheckCommand + precheckScriptFile）。
// 默认未设置 = 拦截生效；显式置 "1"/"true" 才关闭。
func skillVarGuardDisabled() bool {
	v := os.Getenv("DISABLE_EINO_SKILL_VAR_GUARD")
	return v == "1" || v == "true"
}
