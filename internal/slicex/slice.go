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
