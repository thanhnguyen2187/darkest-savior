package dmeta2

import (
	"darkest-savior/dson/dheader"
	"darkest-savior/dson/dmeta1"
	"darkest-savior/dson/lbytes"
	"github.com/pkg/errors"
)

type (
	Entry struct {
		NameHash int `json:"name_hash"`
		Offset   int `json:"offset"`
		// FieldInfo is a 32-bit integer that compact additional information
		//
		//   00 | 0000 0000 0000 0000 0000 | 0 0000 0000 | 01
		//                  ^                   ^           ^
		//                  |                   |           |
		//                  |                   |    last bit is set if is an object
		//                  |                   |
		//                  |            field name length including \0
		//                  |
		//      meta 1 entry index if is an object
		FieldInfo  int        `json:"field_info"`
		Inferences Inferences `json:"inferences"`
	}
	Inferences struct {
		Index             int  `json:"index"`
		IsObject          bool `json:"is_object"`
		ParentIndex       int  `json:"parent_index"`
		FieldNameLength   int  `json:"field_name_length"`
		Meta1EntryIndex   int  `json:"meta_1_entry_index"`
		NumDirectChildren int  `json:"num_direct_children"`
		NumAllChildren    int  `json:"num_all_children"`
		RawDataLength     int  `json:"raw_data_length"`
	}
)

func DecodeEntry(reader *lbytes.Reader) (*Entry, error) {
	readInt := lbytes.CreateIntReadFunction(reader)
	instructions := []lbytes.Instruction{
		{"name_hash", readInt},
		{"offset", readInt},
		{"field_info", readInt},
	}
	meta2Entry, err := lbytes.ExecuteInstructions[Entry](instructions)
	if err != nil {
		err := errors.Wrap(err, "DecodeEntry error")
		return nil, err
	}

	meta2Entry.Inferences = InferUsingFieldInfo(meta2Entry.FieldInfo)

	return meta2Entry, nil
}

func DecodeBlock(reader *lbytes.Reader, header dheader.Header, meta1Blocks []dmeta1.Entry) ([]Entry, error) {
	meta2Entries := make([]Entry, 0, header.NumMeta2Entries)
	for i := 0; i < header.NumMeta2Entries; i++ {
		meta2Entry, err := DecodeEntry(reader)
		if err != nil {
			err := errors.Wrap(err, "DecodeBlock error")
			return nil, err
		}
		if meta2Entry == nil {
			return nil, errors.New("dmeta2.DecodeBlock unreachable code")
		}
		meta2Entries = append(meta2Entries, *meta2Entry)
	}

	meta2Entries = InferIndex(meta2Entries)
	meta2Entries, err := InferRawDataLengths(meta2Entries, header.DataLength)
	if err != nil {
		err := errors.Wrap(err, "dmeta2.DecodeBlock error")
		return nil, err
	}
	meta2Entries, err = InferNumDirectChildren(meta1Blocks, meta2Entries)
	if err != nil {
		err := errors.Wrap(err, "dmeta2.DecodeBlock error")
		return nil, err
	}
	meta2Entries = InferParentIndex(meta2Entries)

	return meta2Entries, nil
}
