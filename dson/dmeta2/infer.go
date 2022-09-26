package dmeta2

import (
	"fmt"

	"darkest-savior/ds"
	"darkest-savior/dson/dmeta1"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

func InferIndex(meta2Entries []Entry) []Entry {
	meta2EntriesCopy := make([]Entry, len(meta2Entries))
	copy(meta2EntriesCopy, meta2Entries)

	for i := range meta2EntriesCopy {
		meta2EntriesCopy[i].Inferences.Index = i
	}
	return meta2EntriesCopy
}

func InferUsingFieldInfo(fieldInfo int32) Inferences {
	inferences := Inferences{
		IsObject:        (fieldInfo & 0b1) == 1,
		FieldNameLength: int((fieldInfo & 0b11111111100) >> 2),
		Meta1EntryIndex: int((fieldInfo & 0b1111111111111111111100000000000) >> 11),
	}

	return inferences
}

func InferRawDataLength(secondOffset int, firstOffset int, firstFieldNameLength int) int {
	return secondOffset - (firstOffset + firstFieldNameLength)
}

func InferParentIndex(meta2Entries []Entry) []Entry {
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

	meta2Entries = ds.BuildTree[Entry](
		// init
		Entry{
			Inferences: Inferences{
				Index:             -1,
				IsObject:          true,
				Meta1EntryIndex:   -1,
				NumDirectChildren: 1,
				RawDataLength:     0,
			},
		},
		// ts
		meta2Entries,
		// popPredicate
		func(entry Entry) bool {
			return entry.Inferences.NumDirectChildren <= 0
		},
		// pushPredicate
		func(entry Entry) bool {
			return entry.Inferences.IsObject &&
				entry.Inferences.NumDirectChildren != 0
		},
		// replaceFunc
		func(entry Entry) Entry {
			entry.Inferences.NumDirectChildren -= 1
			return entry
		},
		// mappingFunc
		func(lastEntry Entry, currentEntry Entry) Entry {
			currentEntry.Inferences.ParentIndex = lastEntry.Inferences.Index
			return currentEntry
		},
	)

	return meta2Entries
}

func InferNumDirectChildren(meta1Entries []dmeta1.Entry, meta2Entries []Entry) ([]Entry, error) {
	// TODO: improve the function by replacing meta1Entries with meta2EntryIndexes
	meta2EntriesCopy := make([]Entry, len(meta2Entries))
	copy(meta2EntriesCopy, meta2Entries)

	for i, meta1Entry := range meta1Entries {
		meta2EntryIndex := meta1Entry.Meta2EntryIndex
		meta2Block := &meta2EntriesCopy[meta2EntryIndex]
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
		meta2Block.Inferences.NumDirectChildren = int(meta1Entry.NumDirectChildren)
		meta2Block.Inferences.NumAllChildren = int(meta1Entry.NumAllChildren)
	}

	return meta2EntriesCopy, nil
}

func InferRawDataLengths(meta2Entries []Entry, headerDataLength int) ([]Entry, error) {
	n := len(meta2Entries)
	// RawDataLength of each meta2Entry is inferred by the difference between
	//
	// - The second block's offset, and
	// - Sum of the first block's offset and the field name length
	//
	// A "normal" loop might work in this case,
	// but at a second glance is not as clear as using `lo.Zip2` to create pairs from the blocks.
	meta2EntriesCopy := lo.Map(
		lo.Zip2(
			meta2Entries[:n-1],
			meta2Entries[1:],
		),
		func(t lo.Tuple2[Entry, Entry], _ int) Entry {
			rawDataLength := InferRawDataLength(
				int(t.B.Offset),
				int(t.A.Offset),
				t.A.Inferences.FieldNameLength,
			)
			t.A.Inferences.RawDataLength = rawDataLength
			return t.A
		},
	)
	meta2Entry, found := lo.Find(
		meta2EntriesCopy,
		func(meta2Block Entry) bool {
			return meta2Block.Inferences.RawDataLength < 0
		},
	)
	if found {
		err := fmt.Errorf(
			`InferRawDataLength meta 2 block "%s" has negative raw data length`,
			ds.DumpJSON(meta2Entry),
		)
		return nil, err
	}

	// lastEntry is an edge case, where within calculation,
	// it is the first entry itself; data length of header serves as the second block's offset
	lastEntry, err := lo.Last(meta2Entries)
	if err != nil {
		msg := fmt.Sprintf("InferRawDataLengths unreachable code where there is no meta 2 block")
		return nil, errors.New(msg)
	}
	lastEntry.Inferences.RawDataLength = InferRawDataLength(
		headerDataLength,
		int(lastEntry.Offset),
		lastEntry.Inferences.FieldNameLength,
	)
	meta2EntriesCopy = append(meta2EntriesCopy, lastEntry)

	return meta2EntriesCopy, nil
}
