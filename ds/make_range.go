package ds

import (
	"golang.org/x/exp/constraints"
)

func MakeRange[T constraints.Ordered](start, end, step T) []T {
	// TODO: optimize by calculating the capacity
	sequence := make([]T, 0)
	for i := start; i < end; i += step {
		sequence = append(sequence, i)
	}
	return sequence
}
