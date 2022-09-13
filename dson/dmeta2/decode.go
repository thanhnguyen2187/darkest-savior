package dmeta2

import (
	"darkest-savior/dson/dheader"
	"darkest-savior/dson/dmeta1"
	"darkest-savior/dson/lbytes"
	"github.com/pkg/errors"
)

type (
	Block struct {
		NameHash   int        `json:"name_hash"`
		Offset     int        `json:"offset"`
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

func DecodeBlock(reader *lbytes.Reader) (*Block, error) {
	readInt := lbytes.CreateIntReadFunction(reader)
	instructions := []lbytes.Instruction{
		{"name_hash", readInt},
		{"offset", readInt},
		{"field_info", readInt},
	}
	meta2Block, err := lbytes.ExecuteInstructions[Block](instructions)
	if err != nil {
		err := errors.Wrap(err, "DecodeBlock error")
		return nil, err
	}

	meta2Block.Inferences = InferUsingFieldInfo(meta2Block.FieldInfo)

	return meta2Block, nil
}

func DecodeBlocks(reader *lbytes.Reader, header dheader.Header, meta1Blocks []dmeta1.Block) ([]Block, error) {
	meta2Blocks := make([]Block, 0, header.NumMeta2Entries)
	for i := 0; i < header.NumMeta2Entries; i++ {
		meta2Block, err := DecodeBlock(reader)
		if err != nil {
			err := errors.Wrap(err, "DecodeBlocks error")
			return nil, err
		}
		if meta2Block == nil {
			return nil, errors.New("dmeta2.DecodeBlocks unreachable code")
		}
		meta2Blocks = append(meta2Blocks, *meta2Block)
	}

	meta2Blocks = InferIndex(meta2Blocks)
	meta2Blocks, err := InferRawDataLengths(meta2Blocks, header.DataLength)
	if err != nil {
		err := errors.Wrap(err, "dmeta2.DecodeBlocks error")
		return nil, err
	}
	meta2Blocks, err = InferNumDirectChildren(meta1Blocks, meta2Blocks)
	if err != nil {
		err := errors.Wrap(err, "dmeta2.DecodeBlocks error")
		return nil, err
	}
	meta2Blocks = InferParentIndex(meta2Blocks)

	return meta2Blocks, nil
}
