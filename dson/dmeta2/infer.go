package dmeta2

import (
	"fmt"

	"darkest-savior/ds"
	"darkest-savior/dson/dmeta1"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

func InferUsingFieldInfo(fieldInfo int) Inferences {
	inferences := Inferences{
		IsObject:        (fieldInfo & 0b1) == 1,
		FieldNameLength: (fieldInfo & 0b11111111100) >> 2,
		Meta1EntryIndex: (fieldInfo & 0b1111111111111111111100000000000) >> 11,
	}

	return inferences
}

func InferRawDataLength(secondOffset int, firstOffset int, firstFieldNameLength int) int {
	return secondOffset - (firstOffset + firstFieldNameLength)
}

func InferParentIndex(meta2Blocks []Block) []Block {
	meta2BlocksCopy := make([]Block, len(meta2Blocks))
	copy(meta2BlocksCopy, meta2Blocks)

	// As the fields in a DSON file are laid sequentially,
	// a stack can be used to find out the parent index of each field.
	//
	// For example, visualizing a stack like this:
	//
	//   [{"index": 0, "num_direct_children": 3}
	//    {"index": 1, "num_direct_children": 0}
	//    {"index": 2, "num_direct_children": 2}
	//    {"index": 3, "num_direct_children": 0}
	//    {"index": 4, "num_direct_children": 0}
	//    {"index": 5, "num_direct_children": 0}]
	//
	// Means the hierarchy looks like this:
	//
	//   0 --> 1
	//    \--> 2 --> 3
	//    |     \--> 4
	//    \--> 5
	type Tracker struct {
		Index             int
		NumDirectChildren int
	}
	stack := ds.NewStack[Tracker]()
	stack.Push(
		Tracker{
			Index:             -1,
			NumDirectChildren: 1,
		},
	) // set a "default" first item to reduce the complexity of edge case handling
	for i := range meta2BlocksCopy {
		meta2Block := &meta2BlocksCopy[i]
		last := stack.Peek()
		meta2Block.Inferences.ParentIndex = last.Index
		stack.ReplaceLast(
			func(oldTracker Tracker) Tracker {
				oldTracker.NumDirectChildren -= 1
				return oldTracker
			},
		)
		last = stack.Peek()
		if last.NumDirectChildren == 0 {
			stack.Pop()
		}
		if meta2Block.Inferences.IsObject {
			stack.Push(
				Tracker{
					Index:             i,
					NumDirectChildren: meta2Block.Inferences.NumDirectChildren,
				},
			)
		}
	}

	return meta2BlocksCopy
}

func InferNumDirectChildren(meta1Blocks []dmeta1.Block, meta2Blocks []Block) ([]Block, error) {
	// TODO: improve the function by replacing meta1Blocks with meta2EntryIndexes
	meta2BlocksCopy := make([]Block, len(meta2Blocks))
	copy(meta2BlocksCopy, meta2Blocks)

	for i, meta1Block := range meta1Blocks {
		meta2EntryIndex := meta1Block.Meta2EntryIndex
		meta2Block := &meta2BlocksCopy[meta2EntryIndex]
		meta1EntryIndex := meta2Block.Inferences.Meta1EntryIndex
		if !meta2Block.Inferences.IsObject {
			err := fmt.Errorf("InferParentIndex metaBlock2 %d is not an object", meta2EntryIndex)
			return nil, err
		}
		if meta1EntryIndex != i {
			err := fmt.Errorf(
				"InferParentIndex invalid meta1EntryIndex of meta2Block %d: expected %d; got %d",
				meta2EntryIndex, i, meta1EntryIndex,
			)
			return nil, err
		}
		meta2Block.Inferences.NumDirectChildren = meta1Block.NumDirectChildren
	}

	return meta2BlocksCopy, nil
}

func InferRawDataLengths(meta2Blocks []Block, headerDataLength int) ([]Block, error) {
	n := len(meta2Blocks)
	// RawDataLength of each meta2Block is inferred by the difference between
	//
	// - The second block's offset, and
	// - Sum of the first block's offset and the field name length
	//
	// A "normal" loop might work in this case,
	// but at a second glance is not as clear as using `lo.Zip2` to create pairs from the blocks.
	meta2BlocksCopy := lo.Map(
		lo.Zip2(
			meta2Blocks[:n-1],
			meta2Blocks[1:],
		),
		func(t lo.Tuple2[Block, Block], _ int) Block {
			rawDataLength := InferRawDataLength(
				t.B.Offset,
				t.A.Offset,
				t.A.Inferences.FieldNameLength,
			)
			t.A.Inferences.RawDataLength = rawDataLength
			return t.A
		},
	)
	meta2Block, found := lo.Find(
		meta2BlocksCopy,
		func(meta2Block Block) bool {
			return meta2Block.Inferences.RawDataLength < 0
		},
	)
	if found {
		err := fmt.Errorf(
			`InferRawDataLength meta 2 block "%s" has negative raw data length`,
			ds.JSONDumps(meta2Block),
		)
		return nil, err
	}

	// lastBlock is an edge case, where within calculation,
	// it is the first block itself; data length of header serves as the second block's offset
	lastBlock, err := lo.Last(meta2Blocks)
	if err != nil {
		msg := fmt.Sprintf("InferRawDataLengths unreachable code where there is no meta 2 block")
		return nil, errors.New(msg)
	}
	lastBlock.Inferences.RawDataLength = InferRawDataLength(
		headerDataLength,
		lastBlock.Offset,
		lastBlock.Inferences.FieldNameLength,
	)
	meta2BlocksCopy = append(meta2BlocksCopy, lastBlock)

	return meta2BlocksCopy, nil
}
