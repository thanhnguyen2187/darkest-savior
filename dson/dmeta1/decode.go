package dmeta1

import (
	"darkest-savior/dson/lbytes"
	"github.com/pkg/errors"
)

type (
	Entry struct {
		ParentIndex       int `json:"parent_index"`
		Meta2EntryIndex   int `json:"meta_2_entry_index"`
		NumDirectChildren int `json:"num_direct_children"`
		NumAllChildren    int `json:"num_all_children"`
	}
)

func DecodeEntry(reader *lbytes.Reader) (*Entry, error) {
	readInt := lbytes.CreateIntReadFunction(reader)

	meta1Instructions := []lbytes.Instruction{
		{"parent_index", readInt},
		{"meta_2_entry_index", readInt},
		{"num_direct_children", readInt},
		{"num_all_children", readInt},
	}
	meta1Entry, err := lbytes.ExecuteInstructions[Entry](meta1Instructions)
	if err != nil {
		err := errors.Wrap(err, "DecodeEntry error")
		return nil, err
	}

	return meta1Entry, nil
}

func DecodeBlock(reader *lbytes.Reader, numMeta1Entries int) ([]Entry, error) {
	meta1Entries := make([]Entry, 0, numMeta1Entries)
	for i := 0; i < numMeta1Entries; i++ {
		meta1Entry, err := DecodeEntry(reader)
		if err != nil {
			err := errors.Wrap(err, "DecodeEntry error")
			return nil, err
		}
		if meta1Entry == nil {
			return nil, errors.New("dmeta1.DecodeBlock unreachable code")
		}
		meta1Entries = append(meta1Entries, *meta1Entry)
	}

	return meta1Entries, nil
}
