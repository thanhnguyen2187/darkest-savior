package dfield

type (
	Field struct {
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
	DataType      string
	EncodingField struct {
		Key           string   `json:"key"`
		ValueType     DataType `json:"value_type"`
		Value         any      `json:"value"`
		Bytes         []byte   `json:"bytes"`
		IsObject      bool     `json:"is_object"`
		HierarchyPath []string `json:"hierarchy_path"`
	}
)

const (
	DataTypeUnknown      = DataType("unknown")
	DataTypeSaveEditor   = DataType("save_editor")
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
)
