package dson

import (
	"fmt"

	"darkest-savior/ds"
	"darkest-savior/dson/dheader"
	"darkest-savior/dson/dmeta1"
	"darkest-savior/dson/dmeta2"
	"darkest-savior/dson/lbytes"
	"github.com/pkg/errors"

	"github.com/samber/lo"
)

type (
	DecodedFile struct {
		Header      dheader.Header `json:"header"`
		Meta1Blocks []dmeta1.Block `json:"meta_1_blocks"`
		Meta2Blocks []dmeta2.Block `json:"meta_2_blocks"`
		Fields      []Field        `json:"fields"`
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

func InferUsingMeta2Block(rawData []byte, meta2block dmeta2.Block) FieldInferences {
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

func DecodeField(reader *lbytes.Reader, meta2Block dmeta2.Block) (*Field, error) {
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

func DecodeFields(reader *lbytes.Reader, meta2Blocks []dmeta2.Block) ([]Field, error) {
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

	header, err := dheader.Decode(reader)
	if err != nil {
		return nil, err
	}
	file.Header = *header
	file.Meta1Blocks, err = dmeta1.DecodeBlocks(reader, header.NumMeta1Entries)
	if err != nil {
		return nil, err
	}

	file.Meta2Blocks, err = dmeta2.DecodeBlocks(reader, file.Header, file.Meta1Blocks)
	if err != nil {
		return nil, err
	}

	file.Fields, err = DecodeFields(reader, file.Meta2Blocks)
	if err != nil {
		return nil, err
	}

	return &file, nil
}
