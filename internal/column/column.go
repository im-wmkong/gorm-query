package column

import (
	"fmt"

	"github.com/im-wmkong/gorm-query/internal/mapx"
	"github.com/im-wmkong/gorm-query/internal/slicex"
	"github.com/spf13/cast"
)

// Value resolves an argument: if it implements fmt.Stringer, use its string form.
func Value[V any](v V) any {
	if e, ok := any(v).(fmt.Stringer); ok {
		return e.String()
	}
	return v
}

// Values resolves a list of arguments.
func Values[V any](args []V) []any {
	return slicex.Map(args, func(v V) any {
		return Value(v)
	})
}

// ValueTo resolves an argument and casts it to the specified basic type.
func ValueTo[T cast.Basic, V any](v V) T {
	return cast.To[T](Value(v))
}

// ValuesTo resolves a list of arguments and casts them to the specified basic type.
func ValuesTo[T cast.Basic, V any](args []V) []T {
	return slicex.Map(args, func(v V) T {
		return ValueTo[T](v)
	})
}

// ToStringMap converts map[K]V to map[string]V.
func ToStringMap[K ~string, V any](values map[K]V) map[string]V {
	return mapx.MapKeys(values, func(k K) string {
		return string(k)
	})
}
