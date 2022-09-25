package dmeta1

import (
	"darkest-savior/dson/lbytes"
)

func EncodeEntry(entry Entry) []byte {
	bs := make([]byte, 0, DefaultEntrySize)
	bs = append(bs, lbytes.EncodeValueInt(entry.ParentIndex)...)
	bs = append(bs, lbytes.EncodeValueInt(entry.Meta2EntryIndex)...)
	bs = append(bs, lbytes.EncodeValueInt(entry.NumDirectChildren)...)
	bs = append(bs, lbytes.EncodeValueInt(entry.NumAllChildren)...)
	return bs
}

func EncodeBlock(meta1Block []Entry) []byte {
	bs := make([]byte, 0, 16*DefaultEntrySize)
	for _, entry := range meta1Block {
		bs = append(bs, EncodeEntry(entry)...)
	}
	return bs
}

func CalculateBlockLength(numEntries int) int {
	return numEntries * DefaultEntrySize
}
