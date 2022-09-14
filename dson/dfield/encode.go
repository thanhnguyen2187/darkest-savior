package dfield

import (
	"encoding/binary"
	"fmt"
	"math"
	"strings"

	"darkest-savior/dson/dhash"
	"darkest-savior/dson/dheader"
	"darkest-savior/dson/lbytes"
	"github.com/samber/lo"
)

type (
	RevisionNotFoundError struct {
		ActualFieldName string
	}
)

func (r RevisionNotFoundError) Error() string {
	msg := fmt.Sprintf(
		`expected "%s" as the first field; got "%s"`,
		FieldNameRevision, r.ActualFieldName,
	)
	return msg
}

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
	case int32:
		valueInt32 := value.(int32)
		valueUInt32 = uint32(valueInt32)
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
	if strings.HasPrefix(valueStr, "###") {
		valueUInt32 := dhash.HashString(valueStr[3:])
		return EncodeValueInt(valueUInt32)
	}
	// +1 to account for the last zero byte
	bs := EncodeValueInt(len(valueStr) + 1)
	bs = append(bs, []byte(valueStr)...)
	bs = append(bs, '\u0000')
	return bs
}

func EncodeValueIntVector(value any) []byte {
	switch value.(type) {
	case []float64:
		valueFloat64Vector := value.([]float64)
		bs := EncodeValueInt(len(valueFloat64Vector))
		bs = lo.Reduce(
			valueFloat64Vector,
			func(bs []byte, valueFloat64 float64, i int) []byte {
				return append(bs, EncodeValueInt(valueFloat64)...)
			},
			bs,
		)
		return bs
	case []string:
		return EncodeValueStringVector(value)
	case []any:
		return EncodeValueHybridVector(value)
	}
	panic("EncodeValueIntVector unreachable code")
}

func EncodeValueFloatVector(value any) []byte {
	valueFloat64Vector := value.([]float64)
	bs := EncodeValueInt(len(valueFloat64Vector))
	bs = lo.Reduce(
		valueFloat64Vector,
		func(bs []byte, valueFloat64 float64, i int) []byte {
			return append(bs, EncodeValueFloat(valueFloat64)...)
		},
		bs,
	)
	return bs
}

func EncodeValueStringVector(value any) []byte {
	valueStringVector, ok := value.([]string)
	if !ok {
		// TODO: find a way to make the encoding process less messy
		return EncodeValueHybridVector(value)
	}
	bs := EncodeValueInt(len(valueStringVector))
	bs = lo.Reduce(
		valueStringVector,
		func(bs []byte, valueStr string, i int) []byte {
			return append(bs, EncodeValueString(valueStr)...)
		},
		bs,
	)
	return bs
}

func EncodeValueHybrid(value any) []byte {
	switch value.(type) {
	case float64:
		return EncodeValueInt(value)
	case string:
		return EncodeValueString(value)
	}
	return nil
}

func EncodeValueHybridVector(value any) []byte {
	valueAnyVector := value.([]any)
	bs := EncodeValueInt(len(valueAnyVector))
	bs = lo.Reduce(
		valueAnyVector,
		func(bs []byte, valueAny any, _ int) []byte {
			return append(bs, EncodeValueHybrid(valueAny)...)
		},
		bs,
	)
	return bs
}

func EncodeValueTwoBool(value any) []byte {
	// TODO: handle error if length is not 2
	valueBoolVector := make([]bool, 2)
	ok := false
	valueBoolVector, ok = value.([]bool)
	if !ok {
		valueHybridVector := value.([]any)
		b0 := valueHybridVector[0].(bool)
		b1 := valueHybridVector[1].(bool)
		valueBoolVector = []bool{b0, b1}
	}
	oneOrZero := func(b bool) byte {
		if b {
			return 1
		} else {
			return 0
		}
	}
	return []byte{
		oneOrZero(valueBoolVector[0]), 0, 0, 0,
		oneOrZero(valueBoolVector[1]), 0, 0, 0,
	}
}

func EncodeValueTwoInt(value any) []byte {
	valueIntVector := value.([]float64)
	return append(
		EncodeValueInt(valueIntVector[0]),
		EncodeValueInt(valueIntVector[1])...,
	)
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
		DataTypeHybridVector: EncodeValueHybridVector,
		DataTypeTwoBool:      EncodeValueTwoBool,
		DataTypeTwoInt:       EncodeValueTwoInt,
		DataTypeFileRaw:      returnNothing,
		DataTypeFileDecoded:  returnNothing,
		DataTypeFileJSON:     returnNothing,
		DataTypeObject:       returnNothing,
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

func EncodeValues(fields []EncodingField) ([]EncodingField, error) {
	err := error(nil)
	fieldsCopy := make([]EncodingField, 0, len(fields))
	for _, field := range fields {
		field.Bytes, err = EncodeValue(field.Key, field.ValueType, field.Value)
		if err != nil {
			return nil, err
		}
		field.IsObject = field.ValueType == DataTypeObject
		fieldsCopy = append(fieldsCopy, field)
	}
	return fieldsCopy, nil
}

func CreateMeta2Block(fields []EncodingField)

func CreateHeader(fields []EncodingField) (*dheader.Header, error) {
	firstField := fields[0]
	if firstField.Key != FieldNameRevision {
		return nil, RevisionNotFoundError{ActualFieldName: firstField.Key}
	}
	meta1Size := lo.CountBy(
		fields[1:],
		func(field EncodingField) bool {
			return field.IsObject
		},
	)
	header := dheader.Header{
		MagicNumber:     dheader.MagicNumberBytes,
		Revision:        firstField.Value.(int),
		HeaderLength:    64,
		Zeroes:          lbytes.CreateZeroBytes(4),
		Meta1Size:       0,
		NumMeta1Entries: meta1Size,
		Meta1Offset:     0,
		Zeroes2:         nil,
		Zeroes3:         nil,
		NumMeta2Entries: 0,
		Meta2Offset:     0,
		Zeroes4:         nil,
		DataLength:      0,
		DataOffset:      0,
	}
	return header, nil
}
