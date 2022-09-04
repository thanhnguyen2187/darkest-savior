package dfield

import (
	"github.com/emirpasic/gods/maps/linkedhashmap"
	"github.com/samber/lo"
)

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
	case []int:
		return DataTypeIntVector
	case []string:
		return DataTypeStringVector
	case []float64:
		return DataTypeFloatVector
	case []any:
		return DataTypeHybridVector
	case linkedhashmap.Map:
		return DataTypeObject
	}
	return DataTypeUnknown
}

func ImplyDataTypeByFieldName(name string) DataType {
	return InferDataTypeByFieldName(name)
}

func FromLinkedHashMap(lhm linkedhashmap.Map) []Field {
	return lo.Map[any, Field](
		lhm.Keys(),
		func(key any, _ int) Field {
			value, _ := lhm.Get(key)
			field := Field{}
			field.Name = key.(string)
			field.Inferences.DataType = ImplyDataTypeByValue(value)

			return field
		},
	)
}
