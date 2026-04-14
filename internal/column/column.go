package column

import (
	"fmt"

	"github.com/im-wmkong/gorm-query/internal/slicex"
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
	return slicex.Map(args, Value)
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
	return slicex.Map(args, ValueTo[T])
}
