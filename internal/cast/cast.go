package cast

// ToStringMap 转换 map[K]V 到 map[string]V
func ToStringMap[K ~string, V any](values map[K]V) map[string]V {
	result := make(map[string]V, len(values))
	for k, v := range values {
		result[string(k)] = v
	}
	return result
}
