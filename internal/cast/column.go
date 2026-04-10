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

// ValueAs 解析参数为指定类型
func ValueAs[T any](v any) T {
	return As[T](Value(v))
}

// ValuesAs 解析参数列表为指定类型
func ValuesAs[T any](args []any) []T {
	return slices.Map(args, ValueAs[T])
}

// As 转换值为指定类型，默认值为零值
func As[T any](value any) T {
	if val, ok := value.(T); ok {
		return val
	}
	var zero T
	return zero
}
