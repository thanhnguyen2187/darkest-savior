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

func PickBytesHeader(fileBytes []byte) []byte {
	return fileBytes[:dheader.DefaultHeaderSize]
}

func PickBytesMeta1Block(numMeta1Entries int, fileBytes []byte) []byte {
	// change PickBytes to use more data from Header
	start := dheader.DefaultHeaderSize
	blockLength := dmeta1.CalculateBlockLength(numMeta1Entries)
	end := start + blockLength
	return fileBytes[start:end]
}

func PickBytesMeta2Block(numMeta1Entries int, numMeta2Entries int, fileBytes []byte) []byte {
	meta1Offset := dheader.DefaultHeaderSize
	meta2Offset := meta1Offset + dmeta1.CalculateBlockLength(numMeta1Entries)
	meta2Length := meta2Offset + numMeta2Entries*dmeta2.DefaultEntrySize

	return fileBytes[meta2Offset:meta2Length]
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
