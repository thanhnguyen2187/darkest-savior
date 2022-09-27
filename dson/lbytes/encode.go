package lbytes

import (
	"encoding/binary"
	"strings"

	"github.com/thanhnguyen2187/darkest-savior/dson/dhash"
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

func EncodeValueString(value any) []byte {
	valueStr := value.(string)
	if strings.HasPrefix(valueStr, "###") {
		valueUInt32 := dhash.HashString(valueStr[3:])
		return EncodeValueInt(valueUInt32)
	}
	// +1 to account for the last zero byte
	bs := EncodeValueInt(len(valueStr) + 1)
	bs = append(bs, []byte(valueStr)...)
	bs = append(bs, '\u0000')
	return bs
}
