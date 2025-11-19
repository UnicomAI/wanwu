package util

import (
	"encoding/json"
	"fmt"
)

/*func json2Str(i interface{}) (string, error) {
	// 序列化为JSON [EN] Serialize to JSON
	jsonData, err := json.Marshal(i)
	if err != nil {
		return "", fmt.Errorf("JSON Marshaling failed: %w", err)
	}
	// 将JSON字节切片转换为字符串并打印 [EN] Convert JSON byte slice to string and print
	jsonString := string(jsonData)
	return jsonString, nil
}*/

func JSONParse[T any](jsonStr string, target *T) error {
	// 解析JSON到目标类型 [EN] Parse JSON into target type
	if err := json.Unmarshal([]byte(jsonStr), target); err != nil {
		return fmt.Errorf("JSON unmarshaling failed: %w", err)
	}
	return nil
}
