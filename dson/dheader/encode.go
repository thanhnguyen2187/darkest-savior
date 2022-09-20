package dheader

import (
	"encoding/binary"
)

func EncodeValueInt(value any) []byte {
	// TODO: move the duplicated code to another place
	valueUInt32 := uint32(0)
	switch value.(type) {
	case float64:
		valueFloat64 := value.(float64)
		valueUInt32 = uint32(valueFloat64)
	case int:
		valueInt := value.(int)
		valueUInt32 = uint32(valueInt)
	case uint32:
		valueUInt32 = value.(uint32)
	case int32:
		valueInt32 := value.(int32)
		valueUInt32 = uint32(valueInt32)
	}
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, valueUInt32)
	return bs
}

func Encode(header Header) []byte {
	bs := make([]byte, 0)
	bs = append(bs, header.MagicNumber...)
	bs = append(bs, EncodeValueInt(header.Revision)...)
	bs = append(bs, EncodeValueInt(header.HeaderLength)...)
	bs = append(bs, header.Zeroes...)
	bs = append(bs, EncodeValueInt(header.Meta1Size)...)
	bs = append(bs, EncodeValueInt(header.NumMeta1Entries)...)
	bs = append(bs, EncodeValueInt(header.Meta1Offset)...)
	bs = append(bs, header.Zeroes2...)
	bs = append(bs, header.Zeroes3...)
	bs = append(bs, EncodeValueInt(header.NumMeta2Entries)...)
	bs = append(bs, EncodeValueInt(header.Meta2Offset)...)
	bs = append(bs, header.Zeroes4...)
	bs = append(bs, EncodeValueInt(header.DataLength)...)
	bs = append(bs, EncodeValueInt(header.DataOffset)...)
	return bs
}
