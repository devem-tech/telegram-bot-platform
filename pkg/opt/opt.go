package opt

// Do executes the provided function fn and returns its result if condition is true.
// If condition is false, it returns the zero value of type T.
func Do[T any](condition bool, fn func() T) T {
	if !condition {
		var zero T

		return zero
	}

	return fn()
}

// Nilable executes the provided function fn and returns its result if the input pointer is not nil.
// If in is nil, it returns the zero value of type Out.
func Nilable[In, Out any](in *In, fn func() Out) Out {
	return Do(in != nil, fn)
}
