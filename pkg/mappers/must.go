package mappers

// MustMapper is a wrapper around Mapper that panics on invalid values instead of returning an error.
type MustMapper[In, Out comparable] struct {
	Mapper[In, Out]
}

// MustMap creates a new MustMapper based on the provided map[Out]In.
//
// MustMapper panics if a lookup fails.
func MustMap[In, Out comparable](outToIn map[Out]In) MustMapper[In, Out] {
	return MustMapper[In, Out]{
		Map(outToIn),
	}
}

// In returns the corresponding value of type In for the given value of type Out.
//
// If the value is not found, it panics.
func (m MustMapper[In, Out]) In(out Out) In {
	res, err := m.Mapper.In(out)
	if err != nil {
		panic(err)
	}

	return res
}

// Out returns the corresponding value of type Out for the given value of type In.
//
// If the value is not found, it panics.
func (m MustMapper[In, Out]) Out(in In) Out {
	res, err := m.Mapper.Out(in)
	if err != nil {
		panic(err)
	}

	return res
}
