package cast

import (
	"fmt"

	"github.com/im-wmkong/gorm-query/internal/slices"
)

// Value 解析参数
func Value(v any) any {
	if e, ok := v.(fmt.Stringer); ok {
		return e.String()
	}
	return v
}

// Values 解析参数列表
func Values(args []any) []any {
	return slices.Map(args, Value)
}

// ValueTo 解析参数为指定类型
func ValueTo[T any](v any) T {
	if val, ok := Value(v).(T); ok {
		return val
	}
	var zero T
	return zero
}

// ValuesTo 解析参数列表为指定类型
func ValuesTo[T any](args []any) []T {
	return slices.Map(args, ValueTo[T])
}

// ToStringMap 转换 map[K]V 到 map[string]V
func ToStringMap[K ~string, V any](values map[K]V) map[string]V {
	result := make(map[string]V, len(values))
	for k, v := range values {
		result[string(k)] = v
	}
	return result
}
