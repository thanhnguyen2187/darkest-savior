package dfield

import (
	"fmt"
)

type (
	DataField struct {
		Name       string     `json:"name"`
		RawData    []byte     `json:"raw_data"`
		Inferences Inferences `json:"inferences"`
	}
	Inferences struct {
		IsObject          bool     `json:"is_object"`
		NumDirectChildren int      `json:"num_direct_children"`
		NumAllChildren    int      `json:"num_all_children"`
		ParentIndex       int      `json:"parent_index"`
		HierarchyPath     []string `json:"hierarchy_path"`
		RawDataOffset     int      `json:"raw_data_offset"`
		RawDataLength     int      `json:"raw_data_length"`
		RawDataStripped   []byte   `json:"raw_data_stripped"`
		DataType          DataType `json:"data_type"`
		Data              any      `json:"data"`
	}
	DataType string

	ErrRevisionNotFound struct {
		Caller          string
		ActualFieldName string
	}
	ErrInvalidDataLength struct {
		Caller   string
		Expected int
		Actual   int
	}
	ErrInvalidDataLengthCustom struct {
		Caller   string
		Expected string
		Actual   int
	}
)

const (
	DataTypeUnknown      = DataType("unknown")
	DataTypeBool         = DataType("bool")
	DataTypeChar         = DataType("char")
	DataTypeInt          = DataType("int")
	DataTypeFloat        = DataType("float")
	DataTypeString       = DataType("string")
	DataTypeIntVector    = DataType("int_vector")
	DataTypeFloatVector  = DataType("float_vector")
	DataTypeStringVector = DataType("string_vector")
	// DataTypeHybridVector denotes a vector that has integer and string mixed together.
	// It seems weird that we need this, but sometimes, the dehashing of integer success partly,
	// which creates a JSON file that has this kind of mixed data.
	DataTypeHybridVector = DataType("hybrid_vector")
	DataTypeTwoBool      = DataType("two_bool")
	DataTypeTwoInt       = DataType("two_int")
	DataTypeFileRaw      = DataType("file_raw")
	DataTypeFileDecoded  = DataType("file_decoded")
	DataTypeFileJSON     = DataType("file_json")
	DataTypeObject       = DataType("object")

	FieldNameRevision = "__revision_dont_touch"
)

func (r ErrRevisionNotFound) Error() string {
	msg := fmt.Sprintf(
		`%s: expected "%s" as the first field; got "%s"`,
		r.Caller, FieldNameRevision, r.ActualFieldName,
	)
	return msg
}

func (r ErrInvalidDataLength) Error() string {
	msg := fmt.Sprintf(
		`%s: expected field length "%d"; got "%d"`,
		r.Caller, r.Expected, r.Actual,
	)
	return msg
}

func (r ErrInvalidDataLengthCustom) Error() string {
	msg := fmt.Sprintf(
		`%s: expected field length "%s"; got "%d"`,
		r.Caller, r.Expected, r.Actual,
	)
	return msg
}
