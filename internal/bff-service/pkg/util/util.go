// @Author wangxm 8/21/星期四 16:01:00 [EN] @Author wangxm 8/21/Thursday 16:01:00
package util

// 拼接 UniqueId [EN] Splice UniqueId
func ConcatAssistantToolUniqueId(typeStr, IdStr string) string {
	if typeStr == "" || IdStr == "" {
		return ""
	}
	return typeStr + "_" + IdStr
}
