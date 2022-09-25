package dfield

import (
	"math"

	"darkest-savior/ds"
	"github.com/iancoleman/orderedmap"
	"github.com/samber/lo"
)

func ImplyDataTypeByFieldName(fieldName string) DataType {
	if IndicateEmbedded(fieldName) {
		return DataTypeFileJSON
	}
	return InferDataTypeByFieldName(fieldName)
}

func ImplyDataTypeByHierarchyPath(hierarchyPath []string) DataType {
	return InferDataTypeByHierarchyPath(hierarchyPath)
}

func ImplyDataTypeByValue(value any) DataType {
	switch value.(type) {
	case bool:
		return DataTypeBool
	case int:
		return DataTypeInt
	case float64:
		valueFloat64 := value.(float64)
		if math.Trunc(valueFloat64) == valueFloat64 {
			return DataTypeInt
		}
		return DataTypeFloat
	case string:
		valueStr := value.(string)
		if len(valueStr) == 1 {
			return DataTypeChar
		}
		return DataTypeString
	case []bool:
		return DataTypeTwoBool
	case []int:
		return DataTypeIntVector
	case []string:
		return DataTypeStringVector
	case []float64:
		return DataTypeFloatVector
	case []any:
		// this edge case is needed since somehow
		// the JSON unmarshalling process can find out that it is array structure with two `any` values,
		// and the two values are parsable to boolean, but in the end it is unable to conclude that:
		// the array is a boolean array
		valueHybridVector := value.([]any)
		if len(valueHybridVector) == 2 {
			_, ok1 := valueHybridVector[0].(bool)
			_, ok2 := valueHybridVector[1].(bool)
			if ok1 && ok2 {
				return DataTypeTwoBool
			}
		}
		return DataTypeHybridVector
	case orderedmap.OrderedMap:
		// valueLhm := value.(orderedmap.OrderedMap)
		// if len(valueLhm.Keys()) == 2 &&
		// 	valueLhm.Keys()[0] == FieldNameRevision &&
		// 	valueLhm.Keys()[1] == FieldNameRoot {
		// 	return DataTypeFileJSON
		// }
		return DataTypeObject
	default:
		return DataTypeUnknown
	}
}

func ImplyDataType(fieldName string, hierarchyPath []string, value any) DataType {
	dataType := DataTypeUnknown
	dataType = ImplyDataTypeByFieldName(fieldName)
	if dataType == DataTypeUnknown {
		dataType = ImplyDataTypeByHierarchyPath(hierarchyPath)
	}
	if dataType == DataTypeUnknown {
		dataType = ImplyDataTypeByValue(value)
	}
	return dataType
}

func CalculateNumDirectChildren(
	lhm orderedmap.OrderedMap,
) []int32 {
	return lo.FlatMap(
		lhm.Keys(),
		func(key string, _ int) []int32 {
			if IndicateEmbedded(key) {
				return []int32{0}
			}
			value, _ := lhm.Get(key)
			valueLhm, ok := value.(orderedmap.OrderedMap)
			if ok {
				return append(
					[]int32{
						int32(len(valueLhm.Keys())),
					},
					CalculateNumDirectChildren(valueLhm)...,
				)
			}
			return []int32{0}
		},
	)
}

func IndicateEmbedded(fieldName string) bool {
	return lo.Contains(
		[]string{"raw_data", "static_save"},
		fieldName,
	)
}

func CalculateNumAllChildren(
	lhm orderedmap.OrderedMap,
) []int32 {
	return lo.FlatMap(
		lhm.Keys(),
		func(key string, _ int) []int32 {
			if IndicateEmbedded(key) {
				return []int32{0}
			}
			value, _ := lhm.Get(key)
			valueLhm, ok := value.(orderedmap.OrderedMap)
			if ok {
				childrenNums := CalculateNumAllChildren(valueLhm)
				return append(
					[]int32{
						int32(len(childrenNums)),
					},
					childrenNums...,
				)
			}
			return []int32{0}
		},
	)
}

func CalculateParentIndexes(
	numsAllChildren []int32,
) []int32 {
	parentIndexes := ds.Repeat(len(numsAllChildren), int32(-1))
	for index, numAllChild := range numsAllChildren {
		copy(parentIndexes[index+1:], ds.Repeat(int(numAllChild), int32(index)))
	}
	return parentIndexes
}

func CalculateMeta1EntryIndexesV2(
	isObjectSlice []bool,
) []int32 {
	indexes := lo.Reduce(
		isObjectSlice,
		func(result []int32, isObject bool, _ int) []int32 {
			last, _ := lo.Last(result)
			if isObject {
				return append(result, last+1)
			} else {
				return append(result, last)
			}
		},
		[]int32{-1},
	)
	indexes = lo.Drop(indexes, 1)
	indexes = lo.Map(
		lo.Zip2(indexes, isObjectSlice),
		func(tuple lo.Tuple2[int32, bool], _ int) int32 {
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
) []int32 {
	return lo.Reduce(
		lo.Zip2(fieldNameLengths, rawDataStrippedLengths),
		func(r []int32, tuple lo.Tuple2[int, int], _ int) []int32 {
			lastOffset, _ := lo.Last(r)
			fieldNameLength := tuple.A
			rawDataLength := tuple.B
			if rawDataLength >= 4 {
				return append(r, int32(ds.NearestDivisibleByM(int(lastOffset)+fieldNameLength, 4)+rawDataLength))
			} else {
				return append(r, lastOffset+int32(fieldNameLength+rawDataLength))
			}
		},
		[]int32{0},
	)
}

func CalculatePaddedBytesCountsV2(
	fieldNameLengths []int,
	meta2Offsets []int32,
	rawDataStrippedLengths []int,
) []int {
	return lo.Map(
		lo.Zip3(fieldNameLengths, meta2Offsets, rawDataStrippedLengths),
		func(tuple lo.Tuple3[int, int32, int], _ int) int {
			fieldNameLength := tuple.A
			meta2Offset := tuple.B
			rawDataStrippedLength := tuple.C

			paddedLength := 0
			if rawDataStrippedLength >= 4 {
				paddedLength = ds.NearestDivisibleByM(int(meta2Offset)+fieldNameLength, 4)
			} else {
				paddedLength = int(meta2Offset) + fieldNameLength
			}
			return paddedLength - (int(meta2Offset) + fieldNameLength)
		},
	)
}
