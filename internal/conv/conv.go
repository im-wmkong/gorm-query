package conv

import "github.com/im-wmkong/gorm-query/internal/mapx"

// ToStringMap 转换 map[K]V 到 map[string]V
func ToStringMap[K ~string, V any](values map[K]V) map[string]V {
	return mapx.MapKeys(values, func(k K) string {
		return string(k)
	})
}
