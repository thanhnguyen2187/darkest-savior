package dheader

func Encode(header Header) []byte {
	bs := make([]byte, 0)
	bs = append(bs, header.MagicNumber...)
}
