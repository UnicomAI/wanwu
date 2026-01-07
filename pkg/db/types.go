package db

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// JSON 映射类型
type JSONMap map[string]interface{}

func (j JSONMap) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSON value")
	}
	return json.Unmarshal(bytes, j)
}

func (JSONMap) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "mysql":
		return "JSON"
	case "postgres":
		return "JSONB"
	default:
		return "TEXT"
	}
}

// LongText 长文本类型
type LongText string

func (l LongText) Value() (driver.Value, error) {
	return string(l), nil
}

func (l *LongText) Scan(value interface{}) error {
	if value == nil {
		*l = ""
		return nil
	}

	switch v := value.(type) {
	case string:
		*l = LongText(v)
	case []byte:
		*l = LongText(v)
	default:
		return fmt.Errorf("failed to scan LongText value: %v", value)
	}
	return nil
}

func (LongText) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "mysql":
		return "LONGTEXT"
	case "postgres":
		return "TEXT"
	default:
		return "TEXT"
	}
}
