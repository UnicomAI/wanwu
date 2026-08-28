package util

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

func IfElse[T any](ok bool, trueValue, falseValue T) T {
	if ok {
		return trueValue
	}
	return falseValue
}

func I64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func MustI64(s string) int64 {
	i64, _ := I64(s)
	return i64
}

func I32(s string) (int32, error) {
	i64, err := I64(s)
	if err != nil {
		return 0, err
	}
	return int32(i64), nil
}

func MustI32(s string) int32 {
	i64, _ := I64(s)
	return int32(i64)
}

func U32(s string) (uint32, error) {
	i64, err := I64(s)
	if err != nil {
		return 0, err
	}
	return uint32(i64), nil
}

func MustU32(s string) uint32 {
	i64, _ := I64(s)
	return uint32(i64)
}

func MustU32s(ss []string) []uint32 {
	ret := make([]uint32, len(ss))
	for i, s := range ss {
		ret[i] = MustU32(s)
	}
	return ret
}

func Int2Str[T ~int | ~int32 | ~uint32 | ~int64](i T) string {
	return strconv.FormatInt(int64(i), 10)
}

// 将 map[string]interface{} 转换为 map[string]string
func ConvertMapToString(req map[string]interface{}) map[string]string {
	formData := make(map[string]string)
	for key, value := range req {
		switch v := value.(type) {
		case string:
			formData[key] = v
		case int, int32, int64, float32, float64, bool:
			formData[key] = fmt.Sprintf("%v", v)
		default:
			// 对于其他类型，可以序列化为JSON字符串或根据需求处理
			if jsonStr, err := json.Marshal(v); err == nil {
				formData[key] = string(jsonStr)
			}
		}
	}
	return formData
}

func CheckAndRemoveBase64Prefix(s string) (string, bool) {
	// 空值直接返回，避免后续处理报错
	if s == "" {
		return "", false
	}
	// 核心判断：是否包含Base64标准前缀分隔符「;base64,」（精准区分Base64和URL）
	const base64PrefixFlag = ";base64,"
	if strings.Contains(s, base64PrefixFlag) {
		// 是带前缀Base64，按逗号分割取纯Base64内容（SplitN避免内容含逗号，最多分割2部分）
		parts := strings.SplitN(s, ",", 2)
		if len(parts) == 2 {
			return parts[1], true
		}
	}
	// 非带前缀Base64（如URL/其他字符串），返回原串+false
	return s, false
}

// GetStringFromMap 从 map 中安全取字符串字段，不存在或类型不匹配时返回空串
func GetStringFromMap(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// GetIntFromMap 从 map 中安全取整数字段，不存在或类型不匹配时返回 nil
func GetIntFromMap(m map[string]interface{}, key string) *int {
	if v, ok := m[key]; ok {
		switch n := v.(type) {
		case int:
			return &n
		case int64:
			i := int(n)
			return &i
		case float64:
			i := int(n)
			return &i
		}
	}
	return nil
}
