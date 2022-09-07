package dfield

import (
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
		return DataTypeHybridVector
	case orderedmap.OrderedMap:
		valueLhm := value.(orderedmap.OrderedMap)
		if len(valueLhm.Keys()) == 1 && valueLhm.Keys()[0] == "base_root" {
			return DataTypeFileJSON
		}
		return DataTypeObject
	}
	return DataTypeUnknown
}

func FromLinkedHashMap(lhm orderedmap.OrderedMap) []EncodingField {
	// TODO: See if unmarshal to float64 is dangerous in the case and find out how to mitigate
	return lo.Flatten(
		// TODO: find a way to handle `lo.Map` with potential error more gracefully
		lo.Map[string, []EncodingField](
			lhm.Keys(),
			func(key string, _ int) []EncodingField {

				field := EncodingField{}
				field.Key = key
				field.HierarchyPath = append(field.HierarchyPath, key)
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
						FromLinkedHashMap(valueLhm)...,
					)
				default:
					field.Value = value
					return []EncodingField{field}
				}
			},
		),
	)
}
