package opt

func Do[T any](condition bool, fn func() T) T {
	if !condition {
		var zero T

		return zero
	}

	return fn()
}

func Nilable[In, Out any](in *In, fn func() Out) Out {
	return Do(in != nil, fn)
}
