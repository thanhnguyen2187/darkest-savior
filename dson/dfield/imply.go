package dfield

import (
	"github.com/iancoleman/orderedmap"
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
	case orderedmap.OrderedMap:
		return DataTypeObject
	}
	return DataTypeUnknown
}

func ImplyDataTypeByFieldName(name string) DataType {
	if name == "base_root" {
		return DataTypeFileJSON
	}
	return InferDataTypeByFieldName(name)
}

// func AttemptImplyNestedFile(dataType DataType, firstKVPair linkedhashmap.Iterator) DataType {
// 	if dataType != DataTypeObject {
// 		return dataType
// 	}
// 	fieldName := firstKVPair.Key().(string)
// 	if fieldName == "" {
// 		return DataTypeFileJSON
// 	}
// 	return dataType
// }

func FromLinkedHashMap(lhm orderedmap.OrderedMap) []Field {
	// TODO: Rethink if a Field is really needed in this case.
	//       What we actually need is some... instruction to write bytes down
	// TODO: See if unmarshal to float64 is dangerous in the case and find out how to mitigate
	return lo.Flatten(
		lo.Map[string, []Field](
			lhm.Keys(),
			func(key string, _ int) []Field {
				field := Field{}

				value, _ := lhm.Get(key)
				fieldName := key
				field.Name = fieldName

				// TODO: check if same logic for inferring needs to be applied since this implementation purely
				//       look at value of each field
				dataType := ImplyDataTypeByValue(value)
				data := value

				field.Inferences.DataType = dataType
				if dataType == DataTypeFileJSON {
					field.Inferences.Data = nil
					return append(
						[]Field{field},
						FromLinkedHashMap(data.(orderedmap.OrderedMap))...,
					)
				} else if dataType == DataTypeObject {
					field.Inferences.Data = nil
					return append(
						[]Field{field},
						FromLinkedHashMap(data.(orderedmap.OrderedMap))...,
					)
				} else {
					field.Inferences.Data = data
					return []Field{field}
				}
			},
		),
	)
}

func FlattenFields([]Field) []Field {
	return nil
}
