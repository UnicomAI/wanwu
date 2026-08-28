package db

import "strings"

var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

// EscapeLike 转义 LIKE 通配符，避免用户输入的 % / _ 被当成通配符导致过滤失效。
// 依赖默认转义符 \，调用方无需写 ESCAPE 子句。
func EscapeLike(s string) string {
	return likeEscaper.Replace(s)
}
