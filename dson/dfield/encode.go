package dfield

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/samber/lo"
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
	valueUInt32 := uint32(0)
	switch value.(type) {
	case float64:
		valueFloat64 := value.(float64)
		valueUInt32 = uint32(valueFloat64)
	case int:
		valueInt := value.(int)
		valueUInt32 = uint32(valueInt)
	case uint32:
		valueUInt32 = value.(uint32)
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
	bs := make([]byte, 4)
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
	bs := EncodeValueInt(len(valueFloat64Vector))
	lo.Reduce(
		valueFloat64Vector,
		func(bs []byte, valueFloat64 float64, i int) []byte {
			return append(bs, EncodeValueInt(valueFloat64)...)
		},
		bs,
	)
	return bs
}

func EncodeValueFloatVector(value any) []byte {
	valueFloat64Vector := value.([]float64)
	bs := EncodeValueInt(len(valueFloat64Vector))
	lo.Reduce(
		valueFloat64Vector,
		func(bs []byte, valueFloat64 float64, i int) []byte {
			return append(bs, EncodeValueFloat(valueFloat64)...)
		},
		bs,
	)
	return bs
}

func EncodeValueStringVector(value any) []byte {
	valueStringVector := value.([]string)
	bs := EncodeValueInt(len(valueStringVector))
	lo.Reduce(
		valueStringVector,
		func(bs []byte, valueStr string, i int) []byte {
			return append(bs, EncodeValueString(valueStr)...)
		},
		bs,
	)
	return bs
}

type ErrNoEncodeFunc struct {
	Key       string
	ValueType DataType
	Value     any
}

func (r ErrNoEncodeFunc) Error() string {
	msg := fmt.Sprintf(
		`no bytes encode function for key "%s", value type "%s", and value "%v"`,
		r.Key, r.ValueType, r.Value,
	)
	return msg
}

func EncodeValue(key string, valueType DataType, value any) ([]byte, error) {
	type EncodeFunc func(any) []byte
	returnNothing := func(any) []byte { return nil }
	dispatchMap := map[DataType]EncodeFunc{
		DataTypeUnknown:      returnNothing,
		DataTypeBool:         EncodeValueBool,
		DataTypeChar:         EncodeValueChar,
		DataTypeInt:          EncodeValueInt,
		DataTypeFloat:        EncodeValueFloat,
		DataTypeString:       EncodeValueString,
		DataTypeIntVector:    EncodeValueIntVector,
		DataTypeFloatVector:  EncodeValueFloatVector,
		DataTypeStringVector: EncodeValueStringVector,
	}
	encodeFunc, ok := dispatchMap[valueType]
	if !ok {
		err := ErrNoEncodeFunc{
			Key:       key,
			ValueType: valueType,
			Value:     value,
		}
		return nil, err
	}
	bs := encodeFunc(value)
	// TODO: see if `Encode` functions need to return error

	return bs, nil
}
