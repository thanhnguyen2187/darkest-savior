package dfield

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/thanhnguyen2187/darkest-savior/ds"
	"github.com/thanhnguyen2187/darkest-savior/dson/dhash"
	"github.com/thanhnguyen2187/darkest-savior/dson/dheader"
	"github.com/thanhnguyen2187/darkest-savior/dson/dmeta2"
	"github.com/thanhnguyen2187/darkest-savior/match"
)

func InferUsingMeta2Entry(rawData []byte, meta2Entry dmeta2.Entry) Inferences {
	rawDataOffset := int(meta2Entry.Offset) + meta2Entry.Inferences.FieldNameLength
	rawDataLength := meta2Entry.Inferences.RawDataLength
	alignedBytesCount := ds.NearestDivisibleByM(rawDataOffset, 4) - rawDataOffset
	rawDataStripped := rawData
	if rawDataLength > alignedBytesCount {
		rawDataStripped = rawData[alignedBytesCount:]
	}

	return Inferences{
		IsObject:          meta2Entry.Inferences.IsObject,
		NumDirectChildren: meta2Entry.Inferences.NumDirectChildren,
		NumAllChildren:    meta2Entry.Inferences.NumAllChildren,
		ParentIndex:       meta2Entry.Inferences.ParentIndex,
		HierarchyPath:     nil,
		RawDataOffset:     rawDataOffset,
		RawDataLength:     rawDataLength,
		RawDataStripped:   rawDataStripped,
	}
}

func InferHierarchyPath(index int, fields []DataField) []string {
	// TODO: create a cache function if there is need for optimization
	fieldName := fields[index].Name
	parentIndex := fields[index].Inferences.ParentIndex
	if parentIndex == -1 {
		return []string{fields[index].Name}
	}
	return append(InferHierarchyPath(parentIndex, fields), fieldName)
}

func InferHierarchyPaths(fields []DataField) []DataField {
	fieldsCopy := lo.Map(
		fields,
		func(t DataField, i int) DataField {
			t.Inferences.HierarchyPath = InferHierarchyPath(i, fields)
			return t
		},
	)
	return fieldsCopy
}

func InferDataTypeByFieldName(fieldName string) DataType {
	switch fieldName {
	case
		"requirement_code":
		return DataTypeChar
	case
		"current_hp",
		"m_Stress":
		return DataTypeFloat
	case
		"read_page_indexes",
		"raid_read_page_indexes",
		"raid_unread_page_indexes",
		"dungeons_unlocked",
		"played_video_list",
		"trinket_retention_ids",
		"last_party_guids",
		"dungeon_history",
		"buff_group_guids",
		"result_event_history",
		"dead_hero_entries",
		"additional_mash_disabled_infestation_monster_class_ids",
		"skill_cooldown_keys",
		"skill_cooldown_values",
		"bufferedSpawningSlotsAvailable",
		"raid_finish_quirk_monster_class_ids",
		"narration_audio_event_queue_tags",
		"dispatched_events":
		return DataTypeIntVector
	case
		"goal_ids",
		"quirk_groups",
		"backgroundNames":
		return DataTypeStringVector
	case
		"killRange":
		return DataTypeTwoInt
	default:
		return DataTypeUnknown
	}
}

func matchOneOfSlices(toBeMatched []any, toMatchSlices [][]any) bool {
	// `match.OneOf` does not work with slice, hence the workaround
	// Also see: https://github.com/alexpantyukhin/go-pattern-match/issues/40
	matcher := match.Match(toBeMatched)
	for _, toMatchSlice := range toMatchSlices {
		matcher.When(toMatchSlice, true)
	}
	matched, _ := matcher.Result()
	return matched
}

func InferDataTypeByHierarchyPath(hierarchyPath []string) DataType {
	hierarchyPathAny := make([]any, 0, len(hierarchyPath))
	for _, item := range hierarchyPath {
		hierarchyPathAny = append(hierarchyPathAny, item)
	}
	switch true {
	case matchOneOfSlices(
		hierarchyPathAny,
		[][]any{
			{"actor", "buff_group", match.ANY, "amount"},
			{"chapters", match.ANY, match.ANY, "percent"},
			{"non_rolled_additional_chances", match.ANY, "chance"},
		},
	):
		return DataTypeFloat
	case matchOneOfSlices(
		hierarchyPathAny,
		[][]any{
			{"mash", "valid_additional_mash_entry_indexes"},
			{"party", "heroes"},
			{"curioGroups", match.ANY, "curios"},
			{"curioGroups", match.ANY, "curio_table_entries"},
			{"backer_heroes", match.ANY, "combat_skills"},
			{"backer_heroes", match.ANY, "camping_skills"},
			{"backer_heroes", match.ANY, "quirks"},
		},
	):
		return DataTypeIntVector
	case matchOneOfSlices(
		hierarchyPathAny,
		[][]any{
			{"roaming_dungeon_2_ids", match.ANY, "s"},
			{"backgroundGroups", match.ANY, "backgrounds"},
			{"backgroundGroups", match.ANY, "background_table_entries"},
		},
	):
		return DataTypeStringVector
	case matchOneOfSlices(
		hierarchyPathAny,
		[][]any{
			{"map", "bounds"},
			{"areas", match.ANY, "bounds"},
			{"areas", match.ANY, "tiles", match.ANY, "mappos"},
			{"areas", match.ANY, "tiles", match.ANY, "sidepos"},
		},
	):
		return DataTypeFloatVector
	}
	return DataTypeUnknown
}

func InferDataTypeByRawData(rawDataStripped []byte) DataType {
	switch true {
	case len(rawDataStripped) == 1:
		b := rawDataStripped[0]
		if 0x20 <= b && b <= 0x7E {
			return DataTypeChar
		} else {
			return DataTypeBool
		}
	case len(rawDataStripped) == 4:
		return DataTypeInt
	case len(rawDataStripped) == 8:
		bs1 := rawDataStripped[:4]
		bs2 := rawDataStripped[4:]
		oneOrZero := func(bs []byte) bool {
			return bytes.Equal(bs, []byte{1, 0, 0, 0}) ||
				bytes.Equal(bs, []byte{0, 0, 0, 0})
		}
		if oneOrZero(bs1) && oneOrZero(bs2) {
			return DataTypeTwoBool
		}
		// `fallthrough` is used since there is an edge case where the string
		// has exactly 8 bytes.
		//
		// To be more specific, in the edge case, after the bytes are not matched,
		// the function "wrongly" thinks that the type is unknown since Golang
		// does not fall through by default.
		fallthrough
	case len(rawDataStripped) >= 8:
		if dheader.IsValidMagicNumber(rawDataStripped[4:8]) {
			return DataTypeFileRaw
		}
		fallthrough
	case len(rawDataStripped) >= 5:
		return DataTypeString
	}
	return DataTypeUnknown
}

func InferDataType(field DataField) DataType {
	if field.Inferences.IsObject {
		return DataTypeObject
	}
	dataType := InferDataTypeByFieldName(field.Name)
	if dataType == DataTypeUnknown {
		hierarchyPathWithoutRoot := field.Inferences.HierarchyPath[1:]
		dataType = InferDataTypeByHierarchyPath(hierarchyPathWithoutRoot)
	}
	if dataType == DataTypeUnknown {
		dataType = InferDataTypeByRawData(field.Inferences.RawDataStripped)
	}
	return dataType
}

func InferDataBool(rawDataStripped []byte) (bool, error) {
	return rawDataStripped[0] == 1, nil
}

func InferDataChar(rawDataStripped []byte) (string, error) {
	lengthExpected := 1
	lengthActual := len(rawDataStripped)
	if lengthExpected != lengthActual {
		return "", ErrInvalidDataLength{
			Caller:   "InferDataBool",
			Expected: lengthExpected,
			Actual:   lengthActual,
		}
	}
	return string(rawDataStripped), nil
}

func InferDataInt(rawDataStripped []byte) (int32, error) {
	lengthExpected := 4
	lengthActual := len(rawDataStripped)
	if lengthExpected != lengthActual {
		return 0, ErrInvalidDataLength{
			Caller:   "InferDataInt",
			Expected: lengthExpected,
			Actual:   lengthActual,
		}
	}
	return int32(binary.LittleEndian.Uint32(rawDataStripped)), nil
}

func InferDataFloat(rawDataStripped []byte) (float32, error) {
	lengthExpected := 4
	lengthActual := len(rawDataStripped)
	if lengthExpected != lengthActual {
		return 0, ErrInvalidDataLength{
			Caller:   "InferDataFloat",
			Expected: lengthExpected,
			Actual:   lengthActual,
		}
	}
	return math.Float32frombits(
		binary.LittleEndian.Uint32(rawDataStripped),
	), nil
}

func InferDataString(rawDataStripped []byte) (string, error) {
	rawLen := len(rawDataStripped)
	if rawLen < 4 {
		err := ErrInvalidDataLengthCustom{
			Caller:   "InferDataString",
			Expected: ">= 4",
			Actual:   len(rawDataStripped),
		}
		return "", err
	}

	strLen, err := InferDataInt(rawDataStripped[:4])
	if err != nil {
		errUC := ds.ErrUnreachableCode{
			Caller: "InferDataString",
		}
		err := multierror.Append(err, errUC)
		return "", err
	}

	str := string(rawDataStripped[4:])
	trueLen := len(str)
	str = str[:len(str)-1]
	if trueLen != int(strLen) {
		err := ErrInvalidDataLength{
			Caller:   "InferDataString",
			Expected: trueLen,
			Actual:   int(strLen),
		}
		return "", err
	}

	return str, nil
}

func InferDataIntVector(rawDataStripped []byte) ([]int32, error) {
	rawLen := len(rawDataStripped)
	if rawLen < 4 && rawLen%4 != 0 {
		err := ErrInvalidDataLengthCustom{
			Caller:   "InferDataIntVector",
			Expected: "divisible by 4",
			Actual:   rawLen,
		}
		return nil, err
	}

	intVectorLen, err := InferDataInt(rawDataStripped[:4])
	if err != nil {
		errUC := ds.ErrUnreachableCode{
			Caller: "InferDataIntVector",
		}
		err := multierror.Append(err, errUC)
		return nil, err
	}

	intVectorByteChunks := ds.MakeChunks(rawDataStripped[4:], 4)
	if len(intVectorByteChunks) != int(intVectorLen) {
		err := ErrInvalidDataLength{
			Caller:   "InferDataIntVector",
			Expected: int(intVectorLen),
			Actual:   len(intVectorByteChunks),
		}
		return nil, err
	}

	intVector := make([]int32, 0, intVectorLen)
	for _, byteChunk := range intVectorByteChunks {
		value, err := InferDataInt(byteChunk)
		if err != nil {
			err := fmt.Errorf(
				`InferDataIntVector unreachable code with input bytes "%v"`,
				rawDataStripped,
			)
			return nil, err
		}
		intVector = append(intVector, value)
	}

	return intVector, nil
}

func InferDataFloatVector(rawDataStripped []byte) ([]float32, error) {
	rawLen := len(rawDataStripped)
	if rawLen < 4 &&
		rawLen%4 != 0 {
		err := ErrInvalidDataLengthCustom{
			Caller:   "InferDataFloatVector",
			Expected: "divisible by 4",
			Actual:   rawLen,
		}
		return nil, err
	}

	floatVectorLen, err := InferDataInt(rawDataStripped[:4])
	if err != nil {
		errUC := ds.ErrUnreachableCode{
			Caller: "InferDataFloatVector",
		}
		err := multierror.Append(err, errUC)
		return nil, err
	}

	floatVectorByteChunks := ds.MakeChunks(rawDataStripped[4:], 4)
	if len(floatVectorByteChunks) != int(floatVectorLen) {
		err := ErrInvalidDataLength{
			Caller:   "InferDataFloatVector",
			Expected: int(floatVectorLen),
			Actual:   len(floatVectorByteChunks),
		}
		return nil, err
	}

	floatVector := make([]float32, 0, floatVectorLen)
	for _, byteChunk := range floatVectorByteChunks {
		value, err := InferDataFloat(byteChunk)
		if err != nil {
			errUC := ds.ErrUnreachableCode{
				Caller: "InferDataIntVector",
			}
			err := multierror.Append(err, errUC)
			return nil, err
		}
		floatVector = append(floatVector, value)
	}

	return floatVector, nil
}

func InferDataStringVector(rawDataStripped []byte) ([]string, error) {
	caller := "InferDataStringVector"
	rawLen := len(rawDataStripped)
	if rawLen < 4 && rawLen%4 != 0 {
		err := ErrInvalidDataLengthCustom{
			Caller:   caller,
			Expected: "divisible by 4",
			Actual:   rawLen,
		}
		return nil, err
	}

	stringVectorLen, err := InferDataInt(rawDataStripped[:4])
	if err != nil {
		errUC := ds.ErrUnreachableCode{
			Caller: caller,
		}
		err := multierror.Append(err, errUC)
		return nil, err
	}

	stringVector := make([]string, 0, stringVectorLen)
	// A cursor is needed to mark the current reading offset,
	// since a DSON string consists of
	//
	// - 4 bytes for length; let us call this n,
	// - the next n - 1 bytes that make up the actual string, and
	// - \u0000 as the terminator.
	cursor := 4
	for i := 0; i < int(stringVectorLen); i++ {
		str, err := InferDataString(rawDataStripped[cursor:])
		if err != nil {
			errUC := ds.ErrUnreachableCode{
				Caller: caller,
			}
			err := multierror.Append(err, errUC)
			return nil, err
		}
		cursor += 4 + len(str) + 1
		stringVector = append(stringVector, str)
	}

	return stringVector, nil
}

func InferDataTwoInt(rawDataStripped []byte) ([]int32, error) {
	// TODO: improve error handling in every functions
	rawLen := len(rawDataStripped)
	if rawLen != 8 {
		err := fmt.Errorf(
			`InferDataTwoInt got invalid input bytes length: expected 8; got "%v" with length %d`,
			rawDataStripped, rawLen,
		)
		return nil, err
	}

	i1, err := InferDataInt(rawDataStripped[:4])
	if err != nil {
		err := fmt.Errorf(
			`InferDataTwoInt first integer unreachable code with input bytes "%v"`,
			rawDataStripped,
		)
		return nil, err
	}
	i2, err := InferDataInt(rawDataStripped[4:])
	if err != nil {
		err := fmt.Errorf(
			`InferDataTwoInt second integer unreachable code with input bytes "%v"`,
			rawDataStripped,
		)
		return nil, err
	}

	return []int32{i1, i2}, nil
}

func InferDataTwoBool(rawDataStripped []byte) ([]bool, error) {
	rawLen := len(rawDataStripped)
	if rawLen != 8 {
		err := fmt.Errorf(
			`InferDataTwoBool got invalid input bytes length: expected 8; got "%v" with length %d`,
			rawDataStripped, rawLen,
		)
		return nil, err
	}

	b1, err := InferDataBool(rawDataStripped[:4])
	if err != nil {
		err := fmt.Errorf(
			`InferDataTwoBool first integer unreachable code with input bytes "%v"`,
			rawDataStripped,
		)
		return nil, err
	}
	b2, err := InferDataBool(rawDataStripped[4:])
	if err != nil {
		err := fmt.Errorf(
			`InferDataTwoBool second integer unreachable code with input bytes "%v"`,
			rawDataStripped,
		)
		return nil, err
	}

	return []bool{b1, b2}, nil
}

func InferData(dataType DataType, rawDataStripped []byte) (any, error) {
	type InferFunc func([]byte) (any, error)
	returnNothing := func([]byte) (any, error) { return nil, nil }
	dispatchMap := map[DataType]InferFunc{
		DataTypeInt:          func(bs []byte) (any, error) { return InferDataInt(bs) },
		DataTypeString:       func(bs []byte) (any, error) { return InferDataString(bs) },
		DataTypeChar:         func(bs []byte) (any, error) { return InferDataChar(bs) },
		DataTypeBool:         func(bs []byte) (any, error) { return InferDataBool(bs) },
		DataTypeFloat:        func(bs []byte) (any, error) { return InferDataFloat(bs) },
		DataTypeIntVector:    func(bs []byte) (any, error) { return InferDataIntVector(bs) },
		DataTypeFloatVector:  func(bs []byte) (any, error) { return InferDataFloatVector(bs) },
		DataTypeStringVector: func(bs []byte) (any, error) { return InferDataStringVector(bs) },
		DataTypeTwoInt:       func(bs []byte) (any, error) { return InferDataTwoInt(bs) },
		DataTypeTwoBool:      func(bs []byte) (any, error) { return InferDataTwoBool(bs) },
		DataTypeFileRaw:      returnNothing,
		DataTypeObject:       returnNothing,
		DataTypeUnknown:      returnNothing,
	}
	inferFunc, ok := dispatchMap[dataType]
	if !ok {
		err := fmt.Errorf(`dfield.InferData could not find relevant infer function for data type "%s"`, dataType)
		return nil, err
	}
	value, err := inferFunc(rawDataStripped)
	if err != nil {
		err := errors.Wrap(err, `dfield.InferData error`)
		return nil, err
	}
	return value, nil
}

func AttemptUnhashInt(field DataField) DataField {
	if field.Inferences.DataType != DataTypeInt {
		return field
	}
	value := field.Inferences.Data.(int32)
	if name, ok := dhash.NameByHash[value]; ok {
		field.Inferences.Data = name
		field.Inferences.DataType = DataTypeString
		return field
	}
	return field
}

func AttemptUnhashIntVector(field DataField) DataField {
	if field.Inferences.DataType != DataTypeIntVector {
		return field
	}
	hashedValues := field.Inferences.Data.([]int32)
	values := make([]any, 0, len(hashedValues))
	convertedCount := 0
	for _, value := range hashedValues {
		if name, ok := dhash.NameByHash[value]; ok {
			values = append(values, name)
			convertedCount += 1
		} else {
			values = append(values, value)
		}
	}
	switch convertedCount {
	case 0:
		return field
	case len(hashedValues):
		field.Inferences.DataType = DataTypeStringVector
		field.Inferences.Data = values
		return field
	default:
		field.Inferences.DataType = DataTypeHybridVector
		field.Inferences.Data = values
		return field
	}
}
