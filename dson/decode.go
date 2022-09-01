package dson

import (
	"fmt"

	"darkest-savior/ds"
	"darkest-savior/dson/dheader"
	"darkest-savior/dson/dmeta1"
	"darkest-savior/dson/lbytes"
	"github.com/pkg/errors"

	"github.com/samber/lo"
)

type (
	DecodedFile struct {
		Header      dheader.Header `json:"header"`
		Meta1Blocks []dmeta1.Block `json:"meta_1_blocks"`
		Meta2Blocks []Meta2Block   `json:"meta_2_blocks"`
		Fields      []Field        `json:"fields"`
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
)

func HashString(s string) int32 {
	return lo.Reduce(
		[]byte(s),
		func(result int32, b byte, _ int) int32 {
			return result*53 + int32(b)
		},
		0,
	)
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

func InferUsingMeta2Block(rawData []byte, meta2block Meta2Block) FieldInferences {
	rawDataOffset := meta2block.Offset + meta2block.Inferences.FieldNameLength
	rawDataLength := meta2block.Inferences.RawDataLength
	alignedBytesCount := ds.NearestDivisibleByM(rawDataOffset, 4) - rawDataOffset
	rawDataStripped := rawData
	if rawDataLength > alignedBytesCount {
		rawDataStripped = rawData[alignedBytesCount:]
	}

	return FieldInferences{
		IsObject:        meta2block.Inferences.IsObject,
		ParentIndex:     meta2block.Inferences.ParentIndex,
		HierarchyPath:   nil,
		RawDataOffset:   rawDataOffset,
		RawDataLength:   rawDataLength,
		RawDataStripped: rawDataStripped,
	}
}

func InferNumDirectChildren(meta1Blocks []dmeta1.Block, meta2Blocks []Meta2Block) ([]Meta2Block, error) {
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

func DecodeMeta2Block(reader *lbytes.Reader) (*Meta2Block, error) {
	readInt := lbytes.CreateIntReadFunction(reader)
	instructions := []lbytes.Instruction{
		{"name_hash", readInt},
		{"offset", readInt},
		{"field_info", readInt},
	}
	meta2Block, err := lbytes.ExecuteInstructions[Meta2Block](instructions)
	if err != nil {
		err := errors.Wrap(err, "DecodeMeta2Block error")
		return nil, err
	}

	meta2Block.Inferences = InferUsingFieldInfo(meta2Block.FieldInfo)

	return meta2Block, nil
}

func DecodeMeta2Blocks(reader *lbytes.Reader, header dheader.Header, meta1Blocks []dmeta1.Block) ([]Meta2Block, error) {
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

func DecodeField(reader *lbytes.Reader, meta2Block Meta2Block) (*Field, error) {
	// manual decoding and mapping is needed since turning data into JSON
	// and parse back does not work for bytes of RawData
	field := Field{}
	err := error(nil)
	readString := lbytes.CreateStringReadFunction(reader, meta2Block.Inferences.FieldNameLength)
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

	field.Inferences = InferUsingMeta2Block(field.RawData, meta2Block)

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

func DecodeFields(reader *lbytes.Reader, meta2Blocks []Meta2Block) ([]Field, error) {
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

func DecodeDSON(reader *lbytes.Reader) (*DecodedFile, error) {
	file := DecodedFile{}
	err := error(nil)

	header, err := dheader.DecodeHeader(reader)
	if err != nil {
		return nil, err
	}
	file.Header = *header
	file.Meta1Blocks, err = dmeta1.DecodeBlocks(reader, header.NumMeta1Entries)
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
