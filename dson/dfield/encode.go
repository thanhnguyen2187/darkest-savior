package dfield

import (
	"encoding/binary"
	"math"
)

func EncodeValueBool(value any) []byte {
	valueBool := value.(bool)
	if valueBool {
		return []byte{1}
	} else {
		return []byte{0}
	}
}

func EncodeValueChar(value any) []byte {
	valueStr := value.(string)
	return []byte{valueStr[0]}
}

func EncodeValueInt(value any) []byte {
	// TODO: research on potential errors of this type of handling
	valueFloat64, ok := value.(float64)
	valueUInt32 := uint32(0)
	if ok {
		valueUInt32 = uint32(valueFloat64)
	} else {
		valueInt := value.(int)
		valueUInt32 = uint32(valueInt)
	}
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, valueUInt32)
	return bs
}

func EncodeValueFloat(value any) []byte {
	// TODO: research on potential errors of this type of handling
	valueFloat64 := value.(float64)
	valueFloat32 := float32(valueFloat64)
	valueUInt32 := math.Float32bits(valueFloat32)
	bs := make([]byte, 0, 4)
	binary.LittleEndian.PutUint32(bs, valueUInt32)
	return bs
}

func EncodeValueString(value any) []byte {
	valueStr := value.(string)
	bs := EncodeValueInt(len(valueStr))
	bs = append(bs, []byte(valueStr)...)
	bs = append(bs, '\u0000')
	return bs
}

func EncodeValueIntVector(value any) []byte {
	valueFloat64Vector := value.([]float64)
}

func EncodeValue(valueType DataType, value any) []byte {
	type EncodeFunc func(any) []byte
	dispatchMap := map[DataType]EncodeFunc{
		DataTypeBool:   EncodeValueBool,
		DataTypeChar:   EncodeValueChar,
		DataTypeInt:    EncodeValueInt,
		DataTypeFloat:  EncodeValueFloat,
		DataTypeString: EncodeValueString,
		// DataTypeIntVector = DataType("int_vector")
		// DataTypeFloatVector = DataType("float_vector")
		// DataTypeStringVector = DataType("string_vector")
	}
	switch valueType {
	case DataTypeUnknown:
		return nil
	case DataTypeBool:
		valueBool := value.(bool)
		if valueBool {
			return []byte{1}
		} else {
			return []byte{0}
		}
	case DataTypeChar:
		valueStr := value.(string)
		return []byte{valueStr[0]}
	case DataTypeInt:
		valueFloat64 := value.(float64)
		valueInt32 := uint32(valueFloat64)
		bs := make([]byte, 0, 4)
		binary.LittleEndian.PutUint32(bs, valueInt32)
		return bs
	case DataTypeFloat:
	}

	return nil
}
