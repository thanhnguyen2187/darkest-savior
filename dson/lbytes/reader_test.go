package lbytes

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

var tests = map[string]struct {
	in  []byte
	out any
}{}

func TestBytesReader_ReadInt(t *testing.T) {
	reader := Reader{
		Reader: *bytes.NewReader(
			[]byte{
				3, 1, 4, 3,
				12, 34, 56, 78,
			},
		),
	}

	resultInt1, err := reader.ReadInt()
	assert.NoError(t, err)
	assert.Equal(t, 50594051, resultInt1)

	resultInt2, err := reader.ReadInt()
	assert.NoError(t, err)
	assert.Equal(t, 1312301580, resultInt2)
}
