package dfield

type (
	DataField struct {
		Name       string     `json:"name"`
		RawData    []byte     `json:"raw_data"`
		Inferences Inferences `json:"inferences"`
	}
	Inferences struct {
		IsObject          bool     `json:"is_object"`
		NumDirectChildren int32    `json:"num_direct_children"`
		NumAllChildren    int32    `json:"num_all_children"`
		ParentIndex       int32    `json:"parent_index"`
		HierarchyPath     []string `json:"hierarchy_path"`
		RawDataOffset     int32    `json:"raw_data_offset"`
		RawDataLength     int32    `json:"raw_data_length"`
		RawDataStripped   []byte   `json:"raw_data_stripped"`
		DataType          DataType `json:"data_type"`
		Data              any      `json:"data"`
	}
	DataType      string
	EncodingField struct {
		Index             int32    `json:"index"`
		Key               string   `json:"key"`
		ValueType         DataType `json:"value_type"`
		Value             any      `json:"value"`
		Bytes             []byte   `json:"bytes"`
		PaddedBytesCount  int32    `json:"padded_zeroes"`
		IsObject          bool     `json:"is_object"`
		ParentIndex       int32    `json:"parent_index"`
		Meta1ParentIndex  int32    `json:"meta1_parent_index"`
		Meta1EntryIndex   int32    `json:"meta1_entry_index"`
		Meta2Offset       int32    `json:"meta2_offset"`
		NumDirectChildren int32    `json:"num_direct_children"`
		NumAllChildren    int32    `json:"num_all_children"`
		HierarchyPath     []string `json:"hierarchy_path"`
	}
	EncodingFieldsWithRevision struct {
		Revision int32           `json:"revision"`
		Fields   []EncodingField `json:"fields"`
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
	FieldNameRoot     = "base_root"
)
