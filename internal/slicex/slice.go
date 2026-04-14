package slicex

func Map[I any, O any](args []I, transform func(I) O) []O {
	if args == nil {
		return nil
	}
	result := make([]O, len(args))
	for i, a := range args {
		result[i] = transform(a)
	}
	return result
}
