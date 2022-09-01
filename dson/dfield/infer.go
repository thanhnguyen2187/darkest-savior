package dfield

import (
	"darkest-savior/ds"
	"darkest-savior/dson/dmeta2"
	"github.com/samber/lo"
)

func InferUsingMeta2Block(rawData []byte, meta2block dmeta2.Block) Inferences {
	rawDataOffset := meta2block.Offset + meta2block.Inferences.FieldNameLength
	rawDataLength := meta2block.Inferences.RawDataLength
	alignedBytesCount := ds.NearestDivisibleByM(rawDataOffset, 4) - rawDataOffset
	rawDataStripped := rawData
	if rawDataLength > alignedBytesCount {
		rawDataStripped = rawData[alignedBytesCount:]
	}

	return Inferences{
		IsObject:        meta2block.Inferences.IsObject,
		ParentIndex:     meta2block.Inferences.ParentIndex,
		HierarchyPath:   nil,
		RawDataOffset:   rawDataOffset,
		RawDataLength:   rawDataLength,
		RawDataStripped: rawDataStripped,
	}
}

func InferHierarchyPath(index int, fields []Field) []string {
	// TODO: create a cache function later
	fieldName := fields[index].Name
	parentIndex := fields[index].Inferences.ParentIndex
	if parentIndex == -1 {
		return []string{fields[index].Name}
	}
	return append(InferHierarchyPath(parentIndex, fields), fieldName)
}

func InferHierarchyPaths(fields []Field) []Field {
	fieldsCopy := lo.Map(
		fields,
		func(t Field, i int) Field {
			t.Inferences.HierarchyPath = InferHierarchyPath(i, fields)
			return t
		},
	)
	return fieldsCopy
}
