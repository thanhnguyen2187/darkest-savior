package dheader

import (
	"bytes"
	"fmt"

	"darkest-savior/dson/lbytes"
	"github.com/pkg/errors"
)

type (
	Header struct {
		MagicNumber     []byte `json:"magic_number"`
		Revision        []byte `json:"revision"`
		HeaderLength    int    `json:"header_length"`
		Zeroes          []byte `json:"zeroes"`
		Meta1Size       int    `json:"meta_1_size"`
		NumMeta1Entries int    `json:"num_meta_1_entries"`
		Meta1Offset     int    `json:"meta_1_offset"`
		Zeroes2         []byte `json:"zeroes_2"`
		Zeroes3         []byte `json:"zeroes_3"`
		NumMeta2Entries int    `json:"num_meta_2_entries"`
		Meta2Offset     int    `json:"meta_2_offset"`
		Zeroes4         []byte `json:"zeroes_4"`
		DataLength      int    `json:"data_length"`
		DataOffset      int    `json:"data_offset"`
	}
)

var MagicNumberBytes = []byte{0x01, 0xB1, 0x00, 0x00}

func createMagicNumberReadFunction(reader *lbytes.Reader) lbytes.ReadFunction {
	return func() (any, error) {
		magicNumberBytes, err := reader.ReadBytes(4)
		if err != nil {
			return nil, err

		}
		if bytes.Compare(magicNumberBytes, MagicNumberBytes) != 0 {
			msg := fmt.Sprintf(
				`invalid magic number: expected "%v", got "%v"`,
				MagicNumberBytes, magicNumberBytes,
			)
			return nil, errors.New(msg)
		}
		return magicNumberBytes, nil
	}
}

func Decode(reader *lbytes.Reader) (*Header, error) {

	readMagicNumber := createMagicNumberReadFunction(reader)
	read4Bytes := lbytes.CreateNBytesReadFunction(reader, 4)
	read8Bytes := lbytes.CreateNBytesReadFunction(reader, 8)
	readInt := lbytes.CreateIntReadFunction(reader)

	headerInstructions := []lbytes.Instruction{
		{"magic_number", readMagicNumber},
		{"revision", read4Bytes},
		{"header_length", readInt},
		{"zeroes", read4Bytes},
		{"meta_1_size", readInt},
		{"num_meta_1_entries", readInt},
		{"meta_1_offset", readInt},
		{"zeroes_2", read8Bytes},
		{"zeroes_3", read8Bytes},
		{"num_meta_2_entries", readInt},
		{"meta_2_offset", readInt},
		{"zeroes_4", read4Bytes},
		{"data_length", readInt},
		{"data_offset", readInt},
	}

	header, err := lbytes.ExecuteInstructions[Header](headerInstructions)
	if err != nil {
		return nil, err
	}

	return header, nil
}
