package slicex

func Map[T, R any](collection []T, transform func(item T) R) []R {
	if collection == nil {
		return nil
	}
	result := make([]R, len(collection))
	for i, item := range collection {
		result[i] = transform(item)
	}
	return result
}
