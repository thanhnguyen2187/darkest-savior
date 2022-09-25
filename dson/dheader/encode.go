package dheader

import (
	"darkest-savior/dson/lbytes"
)

func Encode(header Header) []byte {
	bs := make([]byte, 0, DefaultHeaderSize)
	bs = append(bs, header.MagicNumber...)
	bs = append(bs, lbytes.EncodeValueInt(header.Revision)...)
	bs = append(bs, lbytes.EncodeValueInt(header.HeaderLength)...)
	bs = append(bs, header.Zeroes...)
	bs = append(bs, lbytes.EncodeValueInt(header.Meta1Size)...)
	bs = append(bs, lbytes.EncodeValueInt(header.NumMeta1Entries)...)
	bs = append(bs, lbytes.EncodeValueInt(header.Meta1Offset)...)
	bs = append(bs, header.Zeroes2...)
	bs = append(bs, header.Zeroes3...)
	bs = append(bs, lbytes.EncodeValueInt(header.NumMeta2Entries)...)
	bs = append(bs, lbytes.EncodeValueInt(header.Meta2Offset)...)
	bs = append(bs, header.Zeroes4...)
	bs = append(bs, lbytes.EncodeValueInt(header.DataLength)...)
	bs = append(bs, lbytes.EncodeValueInt(header.DataOffset)...)
	return bs
}
