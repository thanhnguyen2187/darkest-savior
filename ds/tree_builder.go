package ds

import (
	"github.com/samber/lo"
)

// TreeBuilder tries to build a tree from a list presentation,
// assuming that there are the logics on popping and pushing items
// to the underlying stack is provided.
//
// Type T is an information tracker type, while type R is the result data's.
type TreeBuilder[T any] struct {
	stack         Stack[T]
	popPredicate  func(t T) bool
	pushPredicate func(t T) bool
	replaceFunc   func(t T) T
	mappingFunc   func(peekedT T, t T) T
}

func NewTreeBuilder[T any](
	init T,
	popPredicate func(T) bool,
	pushPredicate func(T) bool,
	replaceFunc func(T) T,
	mappingFunc func(T, T) T,
) *TreeBuilder[T] {
	stack := NewStack[T]()
	stack.Push(init)
	return &TreeBuilder[T]{
		stack:         *stack,
		popPredicate:  popPredicate,
		pushPredicate: pushPredicate,
		replaceFunc:   replaceFunc,
		mappingFunc:   mappingFunc,
	}
}

func (r *TreeBuilder[T]) ReplaceLast() {
	last := r.stack.Pop()
	newLast := r.replaceFunc(last)
	r.stack.Push(newLast)
}

func (r *TreeBuilder[T]) Pop() T {
	last := r.stack.Peek()
	if r.popPredicate(last) {
		return r.stack.Pop()
	}
	return last
}

func (r *TreeBuilder[T]) Push(t T) {
	if r.pushPredicate(t) {
		r.stack.Push(t)
	}
}

func (r *TreeBuilder[T]) Peek() T {
	return r.stack.Peek()
}

func (r *TreeBuilder[T]) Build(ts []T) []T {
	tsCopy := make([]T, len(ts))
	copy(tsCopy, ts)
	for i := range tsCopy {
		r.ReplaceLast()
		last := r.Pop()
		tsCopy[i] = r.mappingFunc(last, tsCopy[i])
		r.Push(tsCopy[i])
	}
	return tsCopy
}

func BuildTree[T any](
	initT T,
	ts []T,
	popPredicate func(T) bool,
	pushPredicate func(T) bool,
	replaceFunc func(T) T,
	mappingFunc func(peekedT T, t T) T,
) []T {
	stack := NewStack[T]()
	stack.Push(initT)

	result := lo.Map(
		ts,
		func(t T, _ int) T {
			newT := mappingFunc(stack.Peek(), t)
			stack.ReplaceLast(replaceFunc)
			if popPredicate(stack.Peek()) {
				stack.Pop()
			}
			if pushPredicate(newT) {
				stack.Push(newT)
			}
			return newT
		},
	)

	return result
}
