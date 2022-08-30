package ds

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStack_Peek(t *testing.T) {
	type T struct {
		Value1 int
		Value2 int
	}
	stack := NewStack[T]()
	stack.Push(
		T{
			Value1: 1,
			Value2: 2,
		},
	)

	last := stack.Peek()

	assert.Equal(t, last.Value1, 1)
	assert.Equal(t, last.Value2, 2)
}
