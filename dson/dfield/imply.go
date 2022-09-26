package dfield

import (
	"math"

	"github.com/iancoleman/orderedmap"
	"github.com/samber/lo"
)

func IndicateEmbedded(fieldName string) bool {
	return lo.Contains(
		[]string{"raw_data", "static_save"},
		fieldName,
	)
}

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
		// This edge case is needed since somehow the JSON unmarshalling process finds out that it is array structure
		// with two `any` values, and the two values are parsable to boolean, but in the end it is unable to conclude
		// that the array is a boolean array.
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
