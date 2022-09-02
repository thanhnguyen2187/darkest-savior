package ds

// MakeChunks try to group elements within a slice into smaller "chunk",
// each contains n elements. For example,
//
//   MakeChunks([]int{1, 2, 3, 4, 5}, 2)
//
// should return this exact value:
//
//   [][]int{{1, 2}, {3, 4}, {5, 6}}
func MakeChunks[T any](ts []T, n int) [][]T {
	chunks := make([][]T, 0, len(ts)/n+1)
	for i := range ts {
		if i%n == 0 {
			chunk := ts[i : i+n]
			chunks = append(chunks, chunk)
		}
	}
	return chunks
}
