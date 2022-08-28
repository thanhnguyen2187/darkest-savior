package ds

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLinkedHashMap_Keys(t *testing.T) {
	lhm := NewLinkedHashMap[string, int]()

	assert.True(t, len(lhm.Keys()) == 0)

	lhm.Put("a", 1)
	lhm.Put("b", 2)
	lhm.Put("a", 1)

	assert.Equal(t, []string{"a", "b"}, lhm.Keys())
}

func TestLinkedHashMap_Put(t *testing.T) {
	lhm := NewLinkedHashMap[string, any]()
	lhm.Put("abc", 1)
	lhm.Put("abc", 1)

	assert.Equal(t, lhm.hashMap, map[any]any{"abc": 1})
}

func TestLinkedHashMap_ToJSON(t *testing.T) {
	lhm := NewLinkedHashMap[string, any]()
	lhm.Put("abc", 1)
	lhm.Put("def", 2)

	bs, err := lhm.ToJSON()
	assert.NoError(t, err)

	assert.Equal(t, bs, []byte(`{"abc":1,"def":2}`))
}
