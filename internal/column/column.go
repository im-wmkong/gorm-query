package column

import (
	"fmt"

	"github.com/im-wmkong/gorm-query/internal/mapx"
	"github.com/im-wmkong/gorm-query/internal/slicex"
	"github.com/spf13/cast"
)

// Value resolves an argument: if it implements fmt.Stringer, use its string form.
func Value(v any) any {
	if e, ok := v.(fmt.Stringer); ok {
		return e.String()
	}
	return v
}

// Values resolves a list of arguments.
func Values(args []any) []any {
	return slicex.Map(args, Value)
}

// ValueTo resolves an argument and casts it to the specified basic type.
func ValueTo[T cast.Basic](v any) T {
	return cast.To[T](Value(v))
}

// ValuesTo resolves a list of arguments and casts them to the specified basic type.
func ValuesTo[T cast.Basic](args []any) []T {
	return slicex.Map(args, ValueTo[T])
}

// ToStringMap converts map[K]V to map[string]V.
func ToStringMap[K ~string, V any](values map[K]V) map[string]V {
	return mapx.MapKeys(values, func(k K) string {
		return string(k)
	})
}
