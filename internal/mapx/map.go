package mapx

func MapKeys[K comparable, V any, R comparable](kv map[K]V, iteratee func(key K) R) map[R]V {
	if kv == nil {
		return nil
	}
	result := make(map[R]V, len(kv))
	for k, v := range kv {
		result[iteratee(k)] = v
	}
	return result
}
