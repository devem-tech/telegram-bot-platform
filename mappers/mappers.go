package mappers

import (
	"errors"
	"fmt"
)

var ErrInvalidValue = errors.New("invalid value")

type Mapper[In, Out comparable] struct {
	outToIn map[Out]In
	inToOut map[In]Out
}

func Map[In, Out comparable](outToIn map[Out]In) Mapper[In, Out] {
	inToOut := make(map[In]Out, len(outToIn))
	for out, in := range outToIn {
		inToOut[in] = out
	}

	return Mapper[In, Out]{
		outToIn: outToIn,
		inToOut: inToOut,
	}
}

func (m Mapper[In, Out]) In(out Out) (In, error) {
	res, ok := m.outToIn[out]
	if !ok {
		return res, fmt.Errorf("%w: %v", ErrInvalidValue, out)
	}

	return res, nil
}

func (m Mapper[In, Out]) Out(in In) (Out, error) {
	res, ok := m.inToOut[in]
	if !ok {
		return res, fmt.Errorf("%w: %v", ErrInvalidValue, in)
	}

	return res, nil
}
