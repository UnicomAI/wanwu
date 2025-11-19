package util

import (
	"encoding/json"
	"fmt"
)

/*func json2Str(i interface{}) (string, error) {
	// Serialize to JSON
	jsonData, err := json.Marshal(i)
	if err != nil {
		return "", fmt.Errorf("JSON Marshaling failed: %w", err)
	}
	// Convert JSON byte slice to string and print
	jsonString := string(jsonData)
	return jsonString, nil
}*/

func JSONParse[T any](jsonStr string, target *T) error {
	// Parse JSON into target type
	if err := json.Unmarshal([]byte(jsonStr), target); err != nil {
		return fmt.Errorf("JSON unmarshaling failed: %w", err)
	}
	return nil
}
