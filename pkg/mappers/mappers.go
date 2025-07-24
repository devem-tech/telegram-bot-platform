package mappers

import (
	"errors"
	"fmt"
)

// ErrInvalidValue is returned when a value is not found in the mapping.
var ErrInvalidValue = errors.New("invalid value")

// Mapper provides a bidirectional mapping between types In and Out.
type Mapper[In, Out comparable] struct {
	outToIn map[Out]In
	inToOut map[In]Out
}

// Map creates a new Mapper instance based on the given map[Out]In.
//
// The outToIn parameter is a map that maps values of type Out to values of type In.
// Returns a Mapper that supports bidirectional lookup.
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

// In returns the corresponding value of type In for the given value of type Out.
//
// If the value is not found, ErrInvalidValue is returned.
func (m Mapper[In, Out]) In(out Out) (In, error) {
	res, ok := m.outToIn[out]
	if !ok {
		return res, fmt.Errorf("%w: %v", ErrInvalidValue, out)
	}

	return res, nil
}

// Out returns the corresponding value of type Out for the given value of type In.
//
// If the value is not found, ErrInvalidValue is returned.
func (m Mapper[In, Out]) Out(in In) (Out, error) {
	res, ok := m.inToOut[in]
	if !ok {
		return res, fmt.Errorf("%w: %v", ErrInvalidValue, in)
	}

	return res, nil
}
