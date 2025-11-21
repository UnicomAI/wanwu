// @Author wangxm 8/21/Thursday 16:01:00
package util

// Splice UniqueId
func ConcatAssistantToolUniqueId(typeStr, IdStr string) string {
	if typeStr == "" || IdStr == "" {
		return ""
	}
	return typeStr + "_" + IdStr
}
