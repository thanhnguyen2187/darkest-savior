package dmeta1

type (
	Entry struct {
		ParentIndex       int32 `json:"parent_index"`
		Meta2EntryIndex   int32 `json:"meta_2_entry_index"`
		NumDirectChildren int32 `json:"num_direct_children"`
		NumAllChildren    int32 `json:"num_all_children"`
	}
)

const (
	DefaultEntrySize = 16
)
