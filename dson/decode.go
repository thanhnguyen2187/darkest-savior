package dson

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"

	"github.com/samber/lo"
)

type (
	DecodedFile struct {
		Header      Header       `json:"header"`
		Meta1Blocks []Meta1Block `json:"meta_1_blocks"`
		Meta2Blocks []Meta2Block `json:"meta_2_blocks"`
		Fields      []Field      `json:"fields"`
	}
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
	Meta1Block struct {
		ParentIndex       int `json:"parent_index"`
		Meta2EntryIDx     int `json:"meta_2_entry_idx"`
		NumDirectChildren int `json:"num_direct_children"`
		NumAllChildren    int `json:"num_all_children"`
	}
	Meta2Block struct {
		NameHash   int                  `json:"name_hash"`
		Offset     int                  `json:"offset"`
		FieldInfo  int                  `json:"field_info"`
		Inferences Meta2BlockInferences `json:"inferences"`
	}
	Meta2BlockInferences struct {
		IsPrimitive     bool `json:"is_primitive"`
		FieldNameLength int  `json:"field_name_length"`
		Meta1EntryIndex int  `json:"meta_1_entry_index"`
		RawDataLength   int  `json:"raw_data_length"`
	}
	Field struct {
		Name    string `json:"name"`
		RawData []byte `json:"raw_data"`
	}
	ReadingInstruction struct {
		Key          string
		ReadFunction func() (any, error)
	}
)

var MagicNumberBytes = []byte{0x01, 0xB1, 0x00, 0x00}

func HashString(s string) int32 {
	return lo.Reduce(
		[]byte(s),
		func(result int32, b byte, _ int) int32 {
			return result*53 + int32(b)
		},
		0,
	)
}

func createMagicNumberReadFunction(reader *BytesReader) func() (any, error) {
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

func createNBytesReadFunction(reader *BytesReader, n int) func() (any, error) {
	return func() (any, error) {
		return reader.ReadBytes(n)
	}
}

func createIntReadFunction(reader *BytesReader) func() (any, error) {
	return func() (any, error) {
		return reader.ReadInt()
	}
}

func createStringReadFunction(reader *BytesReader, n int) func() (any, error) {
	return func() (any, error) {
		return reader.ReadString(n)
	}
}

// ExecuteInstructions create the final value t with type T by
//
// - Reading the instruction into a map, then
//
// - Create JSON bytes from the map, and finally
//
// - Read the JSON bytes into t
//
// In order to lessen the burden of manual mapping.
func ExecuteInstructions[T any](instructions []ReadingInstruction) (*T, error) {
	tMap := map[string]any{}
	for _, instruction := range instructions {
		value, err := instruction.ReadFunction()
		if err != nil {
			err := errors.Wrapf(err, `ExecuteInstructions error reading key "%v"`, instruction.Key)
			return nil, err
		}
		tMap[instruction.Key] = value
	}
	tBytes, err := json.Marshal(tMap)
	if err != nil {
		err := errors.Wrapf(err, `ExecuteInstructions error marshalling map "%v" to JSON`, tMap)
		return nil, err
	}

	var t T
	if err := json.Unmarshal(tBytes, &t); err != nil {
		err := errors.Wrapf(
			err, `ExecuteInstructions error unmarshalling bytes "%s" to type "%T"`,
			string(tBytes), t,
		)
		return nil, err
	}

	return &t, nil
}

func DecodeHeader(reader *BytesReader) (*Header, error) {

	readMagicNumber := createMagicNumberReadFunction(reader)
	read4Bytes := createNBytesReadFunction(reader, 4)
	read8Bytes := createNBytesReadFunction(reader, 8)
	readInt := createIntReadFunction(reader)

	headerInstructions := []ReadingInstruction{
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

	header, err := ExecuteInstructions[Header](headerInstructions)
	if err != nil {
		return nil, err
	}

	return header, nil
}

func DecodeMeta1Block(reader *BytesReader) (*Meta1Block, error) {
	readInt := createIntReadFunction(reader)

	meta1Instructions := []ReadingInstruction{
		{"parent_index", readInt},
		{"meta_2_entry_idx", readInt},
		{"num_direct_children", readInt},
		{"num_all_children", readInt},
	}
	meta1Block, err := ExecuteInstructions[Meta1Block](meta1Instructions)
	if err != nil {
		err := errors.Wrap(err, "DecodeMeta1Block error")
		return nil, err
	}

	return meta1Block, nil
}

func DecodeMeta1Blocks(reader *BytesReader, numMeta1Entries int) ([]Meta1Block, error) {
	meta1Blocks := make([]Meta1Block, 0, numMeta1Entries)
	for i := 0; i < numMeta1Entries; i++ {
		meta1Block, err := DecodeMeta1Block(reader)
		if err != nil {
			err := errors.Wrap(err, "DecodeMeta1Block error")
			return nil, err
		}
		if meta1Block == nil {
			return nil, errors.New("DecodeMeta1Blocks unreachable code")
		}
		meta1Blocks = append(meta1Blocks, *meta1Block)
	}

	return meta1Blocks, nil
}

func InferUsingFieldInfo(fieldInfo int) Meta2BlockInferences {
	inferences := Meta2BlockInferences{
		IsPrimitive:     (fieldInfo & 0b1) == 1,
		FieldNameLength: (fieldInfo & 0b11111111100) >> 2,
		Meta1EntryIndex: (fieldInfo & 0b1111111111111111111100000000000) >> 11,
	}

	return inferences
}

func InferRawDataLength(secondOffset int, firstOffset int, firstFieldNameLength int) int {
	return secondOffset - (firstOffset + firstFieldNameLength)
}

func DecodeMeta2Block(reader *BytesReader) (*Meta2Block, error) {
	readInt := createIntReadFunction(reader)
	instructions := []ReadingInstruction{
		{"name_hash", readInt},
		{"offset", readInt},
		{"field_info", readInt},
	}
	meta2Block, err := ExecuteInstructions[Meta2Block](instructions)
	if err != nil {
		err := errors.Wrap(err, "DecodeMeta2Block error")
		return nil, err
	}

	meta2Block.Inferences = InferUsingFieldInfo(meta2Block.FieldInfo)

	return meta2Block, nil
}

func DecodeMeta2Blocks(reader *BytesReader, headerDataLength int, numMeta2Entries int) ([]Meta2Block, error) {
	meta2Blocks := make([]Meta2Block, 0, numMeta2Entries)
	for i := 0; i < numMeta2Entries; i++ {
		meta2Block, err := DecodeMeta2Block(reader)
		if err != nil {
			err := errors.Wrap(err, "DecodeMeta2Blocks error")
			return nil, err
		}
		if meta2Block == nil {
			return nil, errors.New("DecodeMeta2Blocks unreachable code")
		}
		meta2Blocks = append(meta2Blocks, *meta2Block)
	}

	// RawDataLength of each meta2Block is inferred by the difference between
	//
	// - The second block's offset, and
	// - Sum of the first block's offset and the field name length
	//
	// There is an exception for the last block. As there is no second block, header's data length (the whole file's
	// length in bytes) is used as the second block's offset, and field name length is also zero in this case.
	for i := 1; i < numMeta2Entries-1; i++ {
		rawDataLength := InferRawDataLength(
			meta2Blocks[i].Offset,
			meta2Blocks[i-1].Offset,
			meta2Blocks[i-1].Inferences.FieldNameLength,
		)
		if rawDataLength < 0 {
			//goland:noinspection GoErrorStringFormat
			err := fmt.Errorf("DecodeMeta2Blocks cannot have negative raw data length %d", rawDataLength)
			return nil, err
		}
		meta2Blocks[i].Inferences.RawDataLength = InferRawDataLength(
			meta2Blocks[i+1].Offset,
			meta2Blocks[i].Offset,
			meta2Blocks[i].Inferences.FieldNameLength,
		)
	}
	lastBlock := meta2Blocks[numMeta2Entries-1]
	lastBlock.Inferences.RawDataLength = InferRawDataLength(
		lastBlock.Offset,
		headerDataLength,
		0,
	)

	return meta2Blocks, nil
}

func DecodeField(reader *BytesReader, meta2Block Meta2Block) (*Field, error) {

	// manual decoding and mapping is needed since turning data into JSON
	// and parse back does not work for bytes of RawData
	field := Field{}
	err := error(nil)
	field.Name, err = reader.ReadString(meta2Block.Inferences.FieldNameLength)
	if err != nil {
		err := errors.Wrap(err, "DecodeField error")
		return nil, err
	}
	field.RawData, err = reader.ReadBytes(meta2Block.Inferences.RawDataLength)
	if err != nil {
		err := errors.Wrap(err, "DecodeField error")
		return nil, err
	}

	return &field, nil
}

func DecodeFields(reader *BytesReader, meta2Blocks []Meta2Block) ([]Field, error) {
	fields := make([]Field, 0, len(meta2Blocks))
	for _, meta2Block := range meta2Blocks {
		field, err := DecodeField(reader, meta2Block)
		if err != nil {
			err := errors.Wrap(err, "DecodeFields error")
			return nil, err
		}
		fields = append(fields, *field)
	}

	return fields, nil
}

func DecodeDSON(reader *BytesReader) (*DecodedFile, error) {
	file := DecodedFile{}
	err := error(nil)

	header, err := DecodeHeader(reader)
	if err != nil {
		return nil, err
	}
	file.Header = *header
	file.Meta1Blocks, err = DecodeMeta1Blocks(reader, header.NumMeta1Entries)
	if err != nil {
		return nil, err
	}

	file.Meta2Blocks, err = DecodeMeta2Blocks(reader, header.DataLength, header.NumMeta2Entries)
	if err != nil {
		return nil, err
	}

	file.Fields, err = DecodeFields(reader, file.Meta2Blocks)
	if err != nil {
		return nil, err
	}

	return &file, nil
}
