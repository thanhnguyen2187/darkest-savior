package ds

func ShallowCopy[T any](ts []T) []T {
	tsCopy := make([]T, len(ts))
	copy(tsCopy, ts)
	return tsCopy
}
