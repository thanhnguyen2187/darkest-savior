package dson

import (
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestHashString(t *testing.T) {
	expectedValues := map[string]int32{
		"":              0,
		"crusader":      1181166609,
		"plague_doctor": -586237712,
	}

	for s, i := range expectedValues {
		assert.Equal(t, HashString(s), i)
	}
}

func TestInferParentIndex(t *testing.T) {
	createMeta2BlockObject := func(numDirectChildren int) Meta2Block {
		return Meta2Block{
			Inferences: Meta2BlockInferences{
				IsObject:          true,
				NumDirectChildren: numDirectChildren,
			},
		}
	}
	createMeta2BlockPrimitive := func() Meta2Block {
		return Meta2Block{}
	}
	meta2Blocks := []Meta2Block{
		createMeta2BlockObject(3),
		createMeta2BlockPrimitive(),
		createMeta2BlockObject(2),
		createMeta2BlockPrimitive(),
		createMeta2BlockPrimitive(),
		createMeta2BlockPrimitive(),
	}
	meta2Blocks = InferParentIndex(meta2Blocks)
	parentIndexes := lo.Map(
		meta2Blocks,
		func(meta2Block Meta2Block, _ int) int {
			return meta2Block.Inferences.ParentIndex
		},
	)
	assert.Equal(t, []int{-1, 0, 0, 2, 2, 0}, parentIndexes)
}

func TestInferNumDirectChildren(t *testing.T) {
	meta1Blocks := []Meta1Block{
		{
			Meta2EntryIndex:   0,
			NumDirectChildren: 3,
		},
		{
			Meta2EntryIndex:   2,
			NumDirectChildren: 2,
		},
	}
	meta2Blocks := make([]Meta2Block, 6)
	meta2Blocks[0] = Meta2Block{
		Inferences: Meta2BlockInferences{
			IsObject:        true,
			Meta1EntryIndex: 0,
		},
	}
	meta2Blocks[2] = Meta2Block{
		Inferences: Meta2BlockInferences{
			IsObject:        true,
			Meta1EntryIndex: 1,
		},
	}

	meta2Blocks, err := InferNumDirectChildren(meta1Blocks, meta2Blocks)
	assert.NoError(t, err)

	nums := lo.Map(
		meta2Blocks,
		func(meta2Block Meta2Block, _ int) int {
			return meta2Block.Inferences.NumDirectChildren
		},
	)
	assert.Equal(t, []int{3, 0, 2, 0, 0, 0}, nums)
	// TODO: test unhappy case
}

func TestInferHierarchyPath(t *testing.T) {
	createField := func(name string, parentIndex int) Field {
		return Field{
			Name:    name,
			RawData: nil,
			Inferences: FieldInferences{
				ParentIndex: parentIndex,
			},
		}
	}
	fields := []Field{
		createField("0", -1),
		createField("1", 0),
		createField("2", 0),
		createField("3", 2),
		createField("4", 2),
		createField("5", 2),
		createField("6", 3),
		createField("7", 6),
		createField("8", 6),
	}
	resultMap := map[int][]string{
		0: {"0"},
		1: {"0", "1"},
		3: {"0", "2", "3"},
		7: {"0", "2", "3", "6", "7"},
		8: {"0", "2", "3", "6", "8"},
	}

	for input, expected := range resultMap {
		assert.Equal(t, expected, InferHierarchyPath(input, fields))
	}
}
