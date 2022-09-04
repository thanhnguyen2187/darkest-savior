package dfield

import (
	"fmt"

	"darkest-savior/dson/dhash"
	"darkest-savior/dson/dmeta2"
	"darkest-savior/dson/lbytes"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

const (
	DataTypeUnknown         = DataType("unknown")
	DataTypeInt             = DataType("int")
	DataTypeHashedInt       = DataType("hashed_int")
	DataTypeString          = DataType("string")
	DataTypeChar            = DataType("char")
	DataTypeBool            = DataType("bool")
	DataTypeFloat           = DataType("float")
	DataTypeIntVector       = DataType("int_vector")
	DataTypeHashedIntVector = DataType("int_vector")
	DataTypeFloatVector     = DataType("float_vector")
	DataTypeStringVector    = DataType("string_vector")
	DataTypeTwoInt          = DataType("two_int")
	DataTypeTwoBool         = DataType("two_bool")
	DataTypeFileRaw         = DataType("file_raw")
	DataTypeFileDecoded     = DataType("file_decoded")
	DataTypeObject          = DataType("object")
)

type (
	Field struct {
		Name       string     `json:"name"`
		RawData    []byte     `json:"raw_data"`
		Inferences Inferences `json:"inferences"`
	}
	Inferences struct {
		IsObject          bool     `json:"is_object"`
		NumDirectChildren int      `json:"num_direct_children"`
		ParentIndex       int      `json:"parent_index"`
		HierarchyPath     []string `json:"hierarchy_path"`
		RawDataOffset     int      `json:"raw_data_offset"`
		RawDataLength     int      `json:"raw_data_length"`
		RawDataStripped   []byte   `json:"raw_data_stripped"`
		DataType          DataType `json:"data_type"`
		Data              any      `json:"data"`
	}
	DataType string
)

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
	hashed := dhash.HashString(field.Name)
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
	fields = lo.Map(
		fields,
		func(field Field, _ int) Field {
			field.Inferences.DataType = InferDataType(field)
			return field
		},
	)
	for i, field := range fields {
		data, err := InferData(field.Inferences.DataType, field.Inferences.RawDataStripped)
		if err != nil {
			err := errors.Wrap(err, "DecodeFields error")
			return nil, err
		}
		field.Inferences.Data = data
		fields[i] = field
	}

	return fields, nil
}
