package lbytes

func CreateZeroBytes(n int) []byte {
	bs := make([]byte, 0, n)
	for i := 0; i < n; i++ {
		bs = append(bs, 0)
	}
	return bs
}
