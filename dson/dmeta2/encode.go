package dmeta2

import (
	"darkest-savior/dson/lbytes"
)

func EncodeEntry(entry Entry) []byte {
	bs := make([]byte, 0, DefaultEntrySize)
	bs = append(bs, lbytes.EncodeValueInt(entry.NameHash)...)
	bs = append(bs, lbytes.EncodeValueInt(entry.Offset)...)
	bs = append(bs, lbytes.EncodeValueInt(entry.FieldInfo)...)
	return bs
}

func EncodeBlock(meta2Block []Entry) []byte {
	bs := make([]byte, 0, DefaultEntrySize*len(meta2Block))
	for _, entry := range meta2Block {
		entryBytes := EncodeEntry(entry)
		bs = append(bs, entryBytes...)
	}
	return bs
}

func CalculateBlockSize(numEntries int) int {
	return numEntries * DefaultEntrySize
}

func CalculateFieldInfo(
	fieldNameLength int,
	isObject bool,
	meta1EntryIndex int,
) int32 {
	fieldInfo := int32(0)
	if isObject {
		fieldInfo ^= int32(1)
	}
	fieldInfo ^= int32(fieldNameLength << 2)
	fieldInfo ^= int32(meta1EntryIndex << 11)
	return fieldInfo
}
