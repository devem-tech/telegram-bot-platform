package mappers

type MustMapper[In, Out comparable] struct {
	Mapper[In, Out]
}

func MustMap[In, Out comparable](outToIn map[Out]In) MustMapper[In, Out] {
	return MustMapper[In, Out]{
		Map(outToIn),
	}
}

func (m MustMapper[In, Out]) In(out Out) In {
	res, err := m.Mapper.In(out)
	if err != nil {
		panic(err)
	}

	return res
}

func (m MustMapper[In, Out]) Out(in In) Out {
	res, err := m.Mapper.Out(in)
	if err != nil {
		panic(err)
	}

	return res
}
