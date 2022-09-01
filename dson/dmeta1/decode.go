package dmeta1

import (
	"darkest-savior/dson/lbytes"
	"github.com/pkg/errors"
)

type (
	Block struct {
		ParentIndex       int `json:"parent_index"`
		Meta2EntryIndex   int `json:"meta_2_entry_index"`
		NumDirectChildren int `json:"num_direct_children"`
		NumAllChildren    int `json:"num_all_children"`
	}
)

func DecodeBlock(reader *lbytes.Reader) (*Block, error) {
	readInt := lbytes.CreateIntReadFunction(reader)

	meta1Instructions := []lbytes.Instruction{
		{"parent_index", readInt},
		{"meta_2_entry_index", readInt},
		{"num_direct_children", readInt},
		{"num_all_children", readInt},
	}
	meta1Block, err := lbytes.ExecuteInstructions[Block](meta1Instructions)
	if err != nil {
		err := errors.Wrap(err, "DecodeBlock error")
		return nil, err
	}

	return meta1Block, nil
}

func DecodeBlocks(reader *lbytes.Reader, numMeta1Entries int) ([]Block, error) {
	meta1Blocks := make([]Block, 0, numMeta1Entries)
	for i := 0; i < numMeta1Entries; i++ {
		meta1Block, err := DecodeBlock(reader)
		if err != nil {
			err := errors.Wrap(err, "DecodeBlock error")
			return nil, err
		}
		if meta1Block == nil {
			return nil, errors.New("DecodeBlocks unreachable code")
		}
		meta1Blocks = append(meta1Blocks, *meta1Block)
	}

	return meta1Blocks, nil
}
