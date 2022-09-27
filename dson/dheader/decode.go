package dheader

import (
	"bytes"
	"fmt"

	"github.com/pkg/errors"
	"github.com/thanhnguyen2187/darkest-savior/dson/lbytes"
)

func IsValidMagicNumber(bs []byte) bool {
	return bytes.Equal(MagicNumberBytes, bs)
}

func createMagicNumberReadFunction(reader *lbytes.Reader) lbytes.ReadFunction {
	return func() (any, error) {
		bs, err := reader.ReadBytes(4)
		if err != nil {
			return nil, err
		}
		if !IsValidMagicNumber(bs) {
			msg := fmt.Sprintf(
				`invalid magic number: expected "%v", got "%v"`,
				MagicNumberBytes, bs,
			)
			return nil, errors.New(msg)
		}
		return bs, nil
	}
}

func Decode(reader *lbytes.Reader) (*Header, error) {

	readMagicNumber := createMagicNumberReadFunction(reader)
	read4Bytes := lbytes.CreateNBytesReadFunction(reader, 4)
	read8Bytes := lbytes.CreateNBytesReadFunction(reader, 8)
	readInt := lbytes.CreateIntReadFunction(reader)

	headerInstructions := []lbytes.Instruction{
		{"magic_number", readMagicNumber},
		{"revision", readInt},
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
