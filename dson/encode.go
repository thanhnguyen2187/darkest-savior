package dson

import (
	"fmt"

	"darkest-savior/dson/dfield"
	"darkest-savior/dson/dheader"
	"darkest-savior/dson/dmeta1"
	"darkest-savior/dson/dmeta2"
)

type (
	KeyCastError struct {
		Key   any
		Value any
	}
)

func (r KeyCastError) Error() string {
	return fmt.Sprintf(
		`unable to cast key "%v" of value "%v" to string`,
		r.Key, r.Value,
	)
}

func EncodeStruct(file Struct) []byte {
	headerSize := dheader.DefaultHeaderSize
	meta1Size := int(file.Header.Meta1Size)
	meta2Size := dmeta2.CalculateBlockSize(int(file.Header.NumMeta2Entries))
	dataSize := int(file.Header.DataLength)
	totalSize := headerSize + meta1Size + meta2Size + dataSize

	headerBytes := dheader.Encode(file.Header)
	meta1Bytes := dmeta1.EncodeBlock(file.Meta1Block)
	meta2Bytes := dmeta2.EncodeBlock(file.Meta2Block)
	dataFieldsBytes := dfield.EncodeDataFields(file.Fields)

	bs := make([]byte, 0, totalSize)
	bs = append(bs, headerBytes...)
	bs = append(bs, meta1Bytes...)
	bs = append(bs, meta2Bytes...)
	bs = append(bs, dataFieldsBytes...)
	return bs
}
