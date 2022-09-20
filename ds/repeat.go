package ds

func Repeat[T any](n int, initial T) []T {
	ts := make([]T, 0, n)
	for i := 0; i < n; i++ {
		ts = append(ts, initial)
	}
	return ts
}
