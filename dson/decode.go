package dson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"darkest-savior/ds"
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
		Meta2EntryIndex   int `json:"meta_2_entry_index"`
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
		IsObject          bool `json:"is_object"`
		ParentIndex       int  `json:"parent_index"`
		FieldNameLength   int  `json:"field_name_length"`
		Meta1EntryIndex   int  `json:"meta_1_entry_index"`
		NumDirectChildren int  `json:"num_direct_children"`
		RawDataLength     int  `json:"raw_data_length"`
	}
	Field struct {
		Name       string          `json:"name"`
		RawData    []byte          `json:"raw_data"`
		Inferences FieldInferences `json:"inferences"`
	}
	FieldInferences struct {
		IsObject        bool     `json:"is_object"`
		ParentIndex     int      `json:"parent_index"`
		HierarchyPath   []string `json:"hierarchy_path"`
		RawDataOffset   int      `json:"raw_data_offset"`
		RawDataLength   int      `json:"raw_data_length"`
		RawDataStripped []byte   `json:"raw_data_stripped"`
		DataType        string   `json:"data_type"`
		Data            any      `json:"data"`
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
		result, err := reader.ReadString(n)
		if err != nil {
			return "", err
		}
		// zero byte trimming is needed since that is how strings are laid out in a DSON file
		return strings.TrimRight(result, "\u0000"), nil
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
		{"meta_2_entry_index", readInt},
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
		IsObject:        (fieldInfo & 0b1) == 1,
		FieldNameLength: (fieldInfo & 0b11111111100) >> 2,
		Meta1EntryIndex: (fieldInfo & 0b1111111111111111111100000000000) >> 11,
	}

	return inferences
}

func InferRawDataLength(secondOffset int, firstOffset int, firstFieldNameLength int) int {
	return secondOffset - (firstOffset + firstFieldNameLength)
}

func InferUsingMeta2Block(
	inferences FieldInferences,
	rawData []byte,
	meta2block Meta2Block,
) FieldInferences {
	rawDataOffset := meta2block.Offset + meta2block.Inferences.FieldNameLength
	rawDataLength := meta2block.Inferences.RawDataLength
	alignedBytesCount := ds.NearestDivisibleByM(rawDataOffset, 4) - rawDataOffset
	rawDataStripped := rawData
	if rawDataLength > alignedBytesCount {
		rawDataStripped = rawData[alignedBytesCount:]
	}

	inferences.RawDataOffset = rawDataOffset
	inferences.RawDataLength = rawDataLength
	inferences.RawDataStripped = rawDataStripped

	return inferences
}

func InferNumDirectChildren(meta1Blocks []Meta1Block, meta2Blocks []Meta2Block) ([]Meta2Block, error) {
	// TODO: improve the function by replacing meta1Blocks with meta2EntryIndexes
	meta2BlocksCopy := make([]Meta2Block, len(meta2Blocks))
	copy(meta2BlocksCopy, meta2Blocks)

	for i, meta1Block := range meta1Blocks {
		meta2EntryIndex := meta1Block.Meta2EntryIndex
		meta2Block := &meta2BlocksCopy[meta2EntryIndex]
		meta1EntryIndex := meta2Block.Inferences.Meta1EntryIndex
		if !meta2Block.Inferences.IsObject {
			err := fmt.Errorf("InferParentIndex metaBlock2 %d is not an object", meta2EntryIndex)
			return nil, err
		}
		if meta1EntryIndex != i {
			err := fmt.Errorf(
				"InferParentIndex invalid meta1EntryIndex of meta2Block %d: expected %d; got %d",
				meta2EntryIndex, i, meta1EntryIndex,
			)
			return nil, err
		}
		meta2Block.Inferences.NumDirectChildren = meta1Block.NumDirectChildren
	}

	return meta2BlocksCopy, nil
}

func InferParentIndex(meta2Blocks []Meta2Block) []Meta2Block {
	meta2BlocksCopy := make([]Meta2Block, len(meta2Blocks))
	copy(meta2BlocksCopy, meta2Blocks)

	// As the fields in a DSON file are laid sequentially,
	// a stack can be used to find out the parent index of each field.
	//
	// For example, visualizing a stack like this:
	//
	//   [{"index": 0, "num_direct_children": 3}
	//    {"index": 1, "num_direct_children": 0}
	//    {"index": 2, "num_direct_children": 2}
	//    {"index": 3, "num_direct_children": 0}
	//    {"index": 4, "num_direct_children": 0}
	//    {"index": 5, "num_direct_children": 0}]
	//
	// Means the hierarchy looks like this:
	//
	//   0 --> 1
	//    \--> 2 --> 3
	//    |     \--> 4
	//    \--> 5
	type Tracker struct {
		Index             int
		NumDirectChildren int
	}
	stack := ds.NewStack[Tracker]()
	stack.Push(
		Tracker{
			Index:             -1,
			NumDirectChildren: 1,
		},
	) // set a "default" first item to reduce the complexity of edge case handling
	for i := range meta2BlocksCopy {
		meta2Block := &meta2BlocksCopy[i]
		last := stack.Peek()
		meta2Block.Inferences.ParentIndex = last.Index
		stack.ReplaceLast(
			func(oldTracker Tracker) Tracker {
				oldTracker.NumDirectChildren -= 1
				return oldTracker
			},
		)
		last = stack.Peek()
		if last.NumDirectChildren == 0 {
			stack.Pop()
		}
		if meta2Block.Inferences.IsObject {
			stack.Push(
				Tracker{
					Index:             i,
					NumDirectChildren: meta2Block.Inferences.NumDirectChildren,
				},
			)
		}
	}

	return meta2BlocksCopy
}

func InferRawDataLengths(meta2Blocks []Meta2Block, headerDataLength int) ([]Meta2Block, error) {
	n := len(meta2Blocks)
	// RawDataLength of each meta2Block is inferred by the difference between
	//
	// - The second block's offset, and
	// - Sum of the first block's offset and the field name length
	meta2BlocksCopy := lo.Map(
		lo.Zip2(
			meta2Blocks[:n-1],
			meta2Blocks[1:],
		),
		func(t lo.Tuple2[Meta2Block, Meta2Block], _ int) Meta2Block {
			rawDataLength := InferRawDataLength(
				t.B.Offset,
				t.A.Offset,
				t.A.Inferences.FieldNameLength,
			)
			t.A.Inferences.RawDataLength = rawDataLength
			return t.A
		},
	)
	meta2Block, found := lo.Find(
		meta2BlocksCopy,
		func(meta2Block Meta2Block) bool {
			return meta2Block.Inferences.RawDataLength < 0
		},
	)
	if found {
		err := fmt.Errorf(
			`InferRawDataLength meta 2 block "%s" has negative raw data length`,
			ds.JSONDumps(meta2Block),
		)
		return nil, err
	}

	// lastBlock is an edge case, where within calculation,
	// it is the first block itself; data length of header serves as the second block's offset
	lastBlock, err := lo.Last(meta2Blocks)
	if err != nil {
		msg := fmt.Sprintf("InferRawDataLengths unreachable code where there is no meta 2 block")
		return nil, errors.New(msg)
	}
	lastBlock.Inferences.RawDataLength = InferRawDataLength(
		headerDataLength,
		lastBlock.Offset,
		lastBlock.Inferences.FieldNameLength,
	)
	meta2BlocksCopy = append(meta2BlocksCopy, lastBlock)

	return meta2BlocksCopy, nil
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

func DecodeMeta2Blocks(reader *BytesReader, header Header, meta1Blocks []Meta1Block) ([]Meta2Block, error) {
	meta2Blocks := make([]Meta2Block, 0, header.NumMeta2Entries)
	for i := 0; i < header.NumMeta2Entries; i++ {
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

	meta2Blocks, err := InferRawDataLengths(meta2Blocks, header.DataLength)
	if err != nil {
		err := errors.Wrap(err, "DecodeMeta2Blocks error")
		return nil, err
	}
	meta2Blocks, err = InferNumDirectChildren(meta1Blocks, meta2Blocks)
	if err != nil {
		err := errors.Wrap(err, "DecodeMeta2Blocks error")
		return nil, err
	}
	meta2Blocks = InferParentIndex(meta2Blocks)

	return meta2Blocks, nil
}

func DecodeField(reader *BytesReader, meta2Block Meta2Block) (*Field, error) {
	// manual decoding and mapping is needed since turning data into JSON
	// and parse back does not work for bytes of RawData
	field := Field{
		Inferences: FieldInferences{
			IsObject:      meta2Block.Inferences.IsObject,
			ParentIndex:   meta2Block.Inferences.ParentIndex,
			RawDataLength: meta2Block.Inferences.RawDataLength,
			RawDataOffset: meta2Block.Offset + meta2Block.Inferences.FieldNameLength,
		},
	}
	err := error(nil)
	readString := createStringReadFunction(reader, meta2Block.Inferences.FieldNameLength)
	ok := false

	fieldName, err := readString()
	if err != nil {
		err := errors.Wrap(err, "DecodeField error: read field.Name")
		return nil, err
	}
	field.Name, ok = fieldName.(string)
	if !ok {
		err := fmt.Errorf(`DecodeField error: unable to cast value "%v" to string for field.Name`, fieldName)
		return nil, err
	}
	hashed := HashString(field.Name)
	if hashed != int32(meta2Block.NameHash) {
		err := fmt.Errorf(
			`DecodeField error: mismatched hash value of field name "%s"; expected "%d", received "%d"`,
			field.Name, meta2Block.NameHash, hashed,
		)
		return nil, err
	}

	field.RawData, err = reader.ReadBytes(meta2Block.Inferences.RawDataLength)
	if err != nil {
		err := errors.Wrap(err, "DecodeField error: read field.RawData")
		return nil, err
	}

	field.Inferences = InferUsingMeta2Block(
		field.Inferences,
		field.RawData,
		meta2Block,
	)

	return &field, nil
}

func InferHierarchyPath(index int, fields []Field) []string {
	// TODO: create a cache function later
	fieldName := fields[index].Name
	parentIndex := fields[index].Inferences.ParentIndex
	if parentIndex == -1 {
		return []string{fields[index].Name}
	}
	return append(InferHierarchyPath(parentIndex, fields), fieldName)
}

func InferHierarchyPaths(fields []Field) []Field {
	fieldsCopy := lo.Map(
		fields,
		func(t Field, i int) Field {
			t.Inferences.HierarchyPath = InferHierarchyPath(i, fields)
			return t
		},
	)
	return fieldsCopy
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

	fields = InferHierarchyPaths(fields)

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

	file.Meta2Blocks, err = DecodeMeta2Blocks(reader, file.Header, file.Meta1Blocks)
	if err != nil {
		return nil, err
	}

	file.Fields, err = DecodeFields(reader, file.Meta2Blocks)
	if err != nil {
		return nil, err
	}

	return &file, nil
}
