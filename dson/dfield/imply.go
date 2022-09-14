package dfield

import (
	"math"

	"darkest-savior/ds"
	"github.com/iancoleman/orderedmap"
	"github.com/samber/lo"
)

func ImplyDataTypeByFieldName(fieldName string) DataType {
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
		valueLhm := value.(orderedmap.OrderedMap)
		if len(valueLhm.Keys()) == 2 &&
			valueLhm.Keys()[0] == FieldNameRevision &&
			valueLhm.Keys()[1] == FieldNameRoot {
			return DataTypeFileJSON
		}
		return DataTypeObject
	default:
		return DataTypeUnknown
	}
}

func FromLinkedHashMap(
	currentPath []string,
	lhm orderedmap.OrderedMap,
) []EncodingField {
	// TODO: See if unmarshal to float64 is dangerous in the case and find out how to mitigate
	return lo.Flatten(
		// TODO: find a way to handle `lo.Map` with potential error more gracefully
		lo.Map[string, []EncodingField](
			lhm.Keys(),
			func(key string, _ int) []EncodingField {
				field := EncodingField{}
				field.Key = key
				// split to a "special case" and a "normal case" since `base_root` is the...
				// root of a DSON file, and also the root of an embedded file; testing is
				// wrong in the embedded file case without this fix
				if key == FieldNameRoot {
					field.HierarchyPath = []string{FieldNameRoot}
				} else {
					// shallow copy is needed here since a lot of FromLinkedHashMap
					// invocation with the same slice leads to strange errors
					field.HierarchyPath = append(
						ds.ShallowCopy(currentPath),
						key,
					)
				}
				field.ValueType = ImplyDataTypeByFieldName(key)
				if field.ValueType == DataTypeUnknown {
					// skip first item that always include "base_root"
					field.ValueType = ImplyDataTypeByHierarchyPath(field.HierarchyPath[1:])
				}
				value, _ := lhm.Get(key)
				if field.ValueType == DataTypeUnknown {
					field.ValueType = ImplyDataTypeByValue(value)
				}

				switch field.ValueType {
				case DataTypeFileJSON:
					fallthrough
				case DataTypeObject:
					field.Value = nil
					valueLhm := value.(orderedmap.OrderedMap)
					return append(
						[]EncodingField{field},
						FromLinkedHashMap(field.HierarchyPath, valueLhm)...,
					)
				default:
					field.Value = value
					return []EncodingField{field}
				}
			},
		),
	)
}
