package dmeta2

import (
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/thanhnguyen2187/darkest-savior/dson/dmeta1"
)

func createEntryObject(index int, numDirectChildren int) Entry {
	return Entry{
		Inferences: Inferences{
			Index:             index,
			IsObject:          true,
			NumDirectChildren: numDirectChildren,
		},
	}
}

func createEntryPrimitive(index int) Entry {
	return Entry{
		Inferences: Inferences{
			Index:    index,
			IsObject: false,
		},
	}
}

func TestInferParentIndex(t *testing.T) {
	meta2Blocks := []Entry{
		createEntryObject(0, 3),
		createEntryPrimitive(1),
		createEntryObject(2, 2),
		createEntryPrimitive(3),
		createEntryObject(4, 0),
		// createEntryPrimitive(4),
		createEntryPrimitive(5),
	}
	meta2Blocks = InferParentIndex(meta2Blocks)
	parentIndexes := lo.Map(
		meta2Blocks,
		func(meta2Block Entry, _ int) int {
			return meta2Block.Inferences.ParentIndex
		},
	)
	assert.Equal(t, []int{-1, 0, 0, 2, 2, 0}, parentIndexes)
}

func TestInferNumDirectChildren(t *testing.T) {
	meta1Blocks := []dmeta1.Entry{
		{
			Meta2EntryIndex:   0,
			NumDirectChildren: 3,
		},
		{
			Meta2EntryIndex:   2,
			NumDirectChildren: 2,
		},
	}
	meta2Blocks := make([]Entry, 6)
	meta2Blocks[0] = Entry{
		Inferences: Inferences{
			IsObject:        true,
			Meta1EntryIndex: 0,
		},
	}
	meta2Blocks[2] = Entry{
		Inferences: Inferences{
			IsObject:        true,
			Meta1EntryIndex: 1,
		},
	}

	meta2Blocks, err := InferNumDirectChildren(meta1Blocks, meta2Blocks)
	assert.NoError(t, err)

	nums := lo.Map(
		meta2Blocks,
		func(meta2Block Entry, _ int) int {
			return meta2Block.Inferences.NumDirectChildren
		},
	)
	assert.Equal(t, []int{3, 0, 2, 0, 0, 0}, nums)
	// TODO: test unhappy case
}
