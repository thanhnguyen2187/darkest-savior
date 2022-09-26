package dmeta2

import (
	"darkest-savior/dson/dheader"
	"darkest-savior/dson/dmeta1"
	"darkest-savior/dson/lbytes"
	"github.com/pkg/errors"
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
	for i := 0; i < int(header.NumMeta2Entries); i++ {
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
	meta2Entries, err := InferRawDataLengths(meta2Entries, int(header.DataLength))
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
