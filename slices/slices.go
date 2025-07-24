package slices

func Map[T any, R any](collection []T, iteratee func(item T) R) []R {
	res := make([]R, len(collection))

	for i := range collection {
		res[i] = iteratee(collection[i])
	}

	return res
}
