package lbytes

import (
	"bytes"
	"encoding/binary"
)

func NewBytesReader(bs []byte) *Reader {
	return &Reader{
		Reader: *bytes.NewReader(bs),
	}
}

func (b *Reader) ReadInt() (int32, error) {
	bs := make([]byte, 4)
	_, err := b.Read(bs)
	if err != nil {
		return 0, err
	}
	result := binary.LittleEndian.Uint32(bs)
	return int32(result), nil
}

func (b *Reader) ReadLong() (int64, error) {
	bs := make([]byte, 8)
	_, err := b.Read(bs)
	if err != nil {
		return 0, err
	}
	result := binary.LittleEndian.Uint64(bs)
	return int64(result), nil
}

func (b *Reader) ReadBytes(n int) ([]byte, error) {
	bs := make([]byte, n)
	// add return early to avoid EOF error
	// when reader's pointer reach end of file
	// while the number of next bytes to read is 0
	if n == 0 {
		return bs, nil
	}
	_, err := b.Read(bs)
	if err != nil {
		return nil, err
	}
	return bs, nil
}

func (b *Reader) ReadString(n int) (string, error) {
	bs, err := b.ReadBytes(n)
	if err != nil {
		return "", err
	}

	return string(bs), nil
}
