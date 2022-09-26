package dfield

import (
	"darkest-savior/ds"
	"github.com/iancoleman/orderedmap"
	"github.com/samber/lo"
)

func CalculateNumDirectChildren(
	lhm orderedmap.OrderedMap,
) []int {
	return lo.FlatMap(
		lhm.Keys(),
		func(key string, _ int) []int {
			if IndicateEmbedded(key) {
				return []int{0}
			}
			value, _ := lhm.Get(key)
			valueLhm, ok := value.(orderedmap.OrderedMap)
			if ok {
				return append(
					[]int{len(valueLhm.Keys())},
					CalculateNumDirectChildren(valueLhm)...,
				)
			}
			return []int{0}
		},
	)
}

func CalculateNumAllChildren(
	lhm orderedmap.OrderedMap,
) []int {
	return lo.FlatMap(
		lhm.Keys(),
		func(key string, _ int) []int {
			if IndicateEmbedded(key) {
				return []int{0}
			}
			value, _ := lhm.Get(key)
			valueLhm, ok := value.(orderedmap.OrderedMap)
			if ok {
				childrenNums := CalculateNumAllChildren(valueLhm)
				return append(
					[]int{
						int(len(childrenNums)),
					},
					childrenNums...,
				)
			}
			return []int{0}
		},
	)
}

func CalculateParentIndexes(
	numsAllChildren []int,
) []int {
	parentIndexes := ds.Repeat(len(numsAllChildren), int(-1))
	for index, numAllChild := range numsAllChildren {
		copy(parentIndexes[index+1:], ds.Repeat(int(numAllChild), int(index)))
	}
	return parentIndexes
}

func CalculateMeta1EntryIndexesV2(
	isObjectSlice []bool,
) []int {
	indexes := lo.Reduce(
		isObjectSlice,
		func(result []int, isObject bool, _ int) []int {
			last, _ := lo.Last(result)
			if isObject {
				return append(result, last+1)
			} else {
				return append(result, last)
			}
		},
		[]int{-1},
	)
	indexes = lo.Drop(indexes, 1)
	indexes = lo.Map(
		lo.Zip2(indexes, isObjectSlice),
		func(tuple lo.Tuple2[int, bool], _ int) int {
			index := tuple.A
			isObject := tuple.B
			if isObject {
				return index
			}
			return 0
		},
	)
	return indexes
}

func CalculateMeta2OffsetsV2(
	fieldNameLengths []int,
	rawDataStrippedLengths []int,
) []int {
	return lo.Reduce(
		lo.Zip2(fieldNameLengths, rawDataStrippedLengths),
		func(r []int, tuple lo.Tuple2[int, int], _ int) []int {
			lastOffset, _ := lo.Last(r)
			fieldNameLength := tuple.A
			rawDataLength := tuple.B
			if rawDataLength >= 4 {
				return append(r, int(ds.NearestDivisibleByM(int(lastOffset)+fieldNameLength, 4)+rawDataLength))
			} else {
				return append(r, lastOffset+int(fieldNameLength+rawDataLength))
			}
		},
		[]int{0},
	)
}

func CalculatePaddedBytesCountsV2(
	fieldNameLengths []int,
	meta2Offsets []int,
	rawDataStrippedLengths []int,
) []int {
	return lo.Map(
		lo.Zip3(fieldNameLengths, meta2Offsets, rawDataStrippedLengths),
		func(tuple lo.Tuple3[int, int, int], _ int) int {
			fieldNameLength := tuple.A
			meta2Offset := tuple.B
			rawDataStrippedLength := tuple.C

			paddedLength := 0
			if rawDataStrippedLength >= 4 {
				paddedLength = ds.NearestDivisibleByM(meta2Offset+fieldNameLength, 4)
			} else {
				paddedLength = meta2Offset + fieldNameLength
			}
			return paddedLength - (meta2Offset + fieldNameLength)
		},
	)
}
