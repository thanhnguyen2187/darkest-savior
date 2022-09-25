package dheader

type (
	Header struct {
		MagicNumber     []byte `json:"magic_number"`
		Revision        int32  `json:"revision"`
		HeaderLength    int32  `json:"header_length"`
		Zeroes          []byte `json:"zeroes"`
		Meta1Size       int32  `json:"meta_1_size"`
		NumMeta1Entries int32  `json:"num_meta_1_entries"`
		Meta1Offset     int32  `json:"meta_1_offset"`
		Zeroes2         []byte `json:"zeroes_2"`
		Zeroes3         []byte `json:"zeroes_3"`
		NumMeta2Entries int32  `json:"num_meta_2_entries"`
		Meta2Offset     int32  `json:"meta_2_offset"`
		Zeroes4         []byte `json:"zeroes_4"`
		DataLength      int32  `json:"data_length"`
		DataOffset      int32  `json:"data_offset"`
	}
)

const (
	DefaultHeaderSize = 64
)

var (
	MagicNumberBytes = []byte{0x01, 0xB1, 0x00, 0x00}
)
