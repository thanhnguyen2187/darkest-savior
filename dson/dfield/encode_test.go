package dfield

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeValueInt(t *testing.T) {
	bs1 := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs1, 10)
	bs2 := EncodeValueInt(10.0)
	assert.Equal(t, bs1, bs2)
}

func TestEncodeValueString(t *testing.T) {
	bs1Len := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs1Len, 3)
	bs1 := append(
		bs1Len,
		[]byte{'a', 'b', 'c', '\u0000'}...,
	)
	bs2 := EncodeValueString("abc")
	assert.Equal(t, bs1, bs2)
}
