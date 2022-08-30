package ds

type Stack[T any] struct {
	slice []T
}

func NewStack[T any]() *Stack[T] {
	return &Stack[T]{
		slice: make([]T, 0),
	}
}

func (r *Stack[T]) Len() int {
	return len(r.slice)
}

func (r *Stack[T]) Push(t T) T {
	r.slice = append(r.slice, t)
	return t
}

func (r *Stack[T]) Pop() T {
	last := r.slice[r.Len()-1]
	r.slice = r.slice[:r.Len()-1]
	return last
}

func (r *Stack[T]) Peek() T {
	last := r.slice[r.Len()-1]
	return last
}

func (r *Stack[T]) ReplaceLast(replacer func(t T) T) T {
	last := r.slice[r.Len()-1]
	newLast := replacer(last)
	r.slice[r.Len()-1] = newLast
	return newLast
}
