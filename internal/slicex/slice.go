package slicex

func Map[T, R any, Slice ~[]T](collection Slice, transform func(item T) R) []R {
	if collection == nil {
		return nil
	}
	result := make([]R, len(collection))
	for i, item := range collection {
		result[i] = transform(item)
	}
	return result
}

func Filter[T any, Slice ~[]T](collection Slice, predicate func(item T) bool) Slice {
	if collection == nil {
		return nil
	}
	result := make(Slice, 0, len(collection))
	for _, item := range collection {
		if predicate(item) {
			result = append(result, item)
		}
	}
	return result
}

// ToMap reduces a slice into a map by extracting (key, value) from each item.
// When two items produce the same key, the later one overrides the earlier
// one, matching how update assignments are typically merged.
//
// Returns nil for a nil/empty slice so callers can pass the result straight
// into APIs that treat nil as "no fields".
func ToMap[T any, K comparable, V any, Slice ~[]T](collection Slice, kv func(item T) (K, V)) map[K]V {
	if len(collection) == 0 {
		return nil
	}
	result := make(map[K]V, len(collection))
	for _, item := range collection {
		k, v := kv(item)
		result[k] = v
	}
	return result
}
