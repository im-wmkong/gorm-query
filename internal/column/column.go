package column

import (
	"fmt"

	"github.com/im-wmkong/gorm-query/internal/mapx"
	"github.com/im-wmkong/gorm-query/internal/slicex"
	"github.com/spf13/cast"
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
func ValueTo[T cast.Basic](v any) T {
	return cast.To[T](Value(v))
}

// ValuesTo 解析参数列表为指定类型
func ValuesTo[T cast.Basic](args []any) []T {
	return slicex.Map(args, ValueTo[T])
}

// ToStringMap 转换 map[K]V 到 map[string]V
func ToStringMap[K ~string, V any](values map[K]V) map[string]V {
	return mapx.MapKeys(values, func(k K) string {
		return string(k)
	})
}
