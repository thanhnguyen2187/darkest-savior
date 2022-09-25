package dmeta2

type (
	Entry struct {
		NameHash int32 `json:"name_hash"`
		Offset   int32 `json:"offset"`
		// FieldInfo is a 32-bit integer that compact additional information
		//
		//   0 | 0000 0000 0000 0000 0000 | 0 0000 0000 | 01
		//   ^             ^                   ^           ^
		//   |             |                   |           |
		//   |             |                   |    last bit is set if is an object
		//   |             |                   |
		//   |             |            field name length including \0
		//   |             |
		//   |    meta 1 entry index
		//   |    if is an object
		//   |
		//   seems to have no meaning, even though it is set to 1 sometimes
		//   also see: https://github.com/robojumper/DarkestDungeonSaveEditor/issues/50
		FieldInfo  int32      `json:"field_info"`
		Inferences Inferences `json:"inferences"`
	}
	Inferences struct {
		Index             int32 `json:"index"`
		IsObject          bool  `json:"is_object"`
		ParentIndex       int32 `json:"parent_index"`
		FieldNameLength   int32 `json:"field_name_length"`
		Meta1EntryIndex   int32 `json:"meta_1_entry_index"`
		NumDirectChildren int32 `json:"num_direct_children"`
		NumAllChildren    int32 `json:"num_all_children"`
		RawDataLength     int32 `json:"raw_data_length"`
	}
)

const (
	DefaultEntrySize = 12
)
