package util

import (
	"fmt"
	"math"
)

// any2float64 将 JSON 反序列化得到的 any 数字（float64/float32/int 系）转为 float64。
// 非数字类型返回 ok=false。
func any2float64(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	default:
		return 0, false
	}
}

// FloatFieldInRange 校验 JSON 对象 m 中名为 key 的数值字段落在 [min,max] 闭区间。
// 语义为"可选字段"：key 不存在时返回 nil（交由上层填默认值）。
// 存在但非数字、非有限值（NaN/Inf）、或超出范围时返回错误。
func FloatFieldInRange(m map[string]any, key string, min, max float64) error {
	v, ok := m[key]
	if !ok {
		return nil
	}
	f, ok := any2float64(v)
	if !ok {
		return fmt.Errorf("%s must be a number", key)
	}
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return fmt.Errorf("%s must be a finite number", key)
	}
	if f < min || f > max {
		return fmt.Errorf("%s out of range [%v,%v]", key, min, max)
	}
	return nil
}
