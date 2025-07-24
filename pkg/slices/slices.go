package slices

// Map applies the provided function iteratee to each element of the input slice collection,
// returning a new slice with the results.
//
// collection - the input slice of type T
// iteratee - a function that takes an item of type T and returns a value of type R
//
// Returns a new slice of type []R with the mapped values.
func Map[T any, R any](collection []T, iteratee func(item T) R) []R {
	res := make([]R, len(collection))

	for i := range collection {
		res[i] = iteratee(collection[i])
	}

	return res
}
