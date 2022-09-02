package dfield

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"

	"darkest-savior/ds"
	"darkest-savior/dson/dheader"
	"darkest-savior/dson/dmeta2"
	"darkest-savior/match"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

func InferUsingMeta2Block(rawData []byte, meta2block dmeta2.Block) Inferences {
	rawDataOffset := meta2block.Offset + meta2block.Inferences.FieldNameLength
	rawDataLength := meta2block.Inferences.RawDataLength
	alignedBytesCount := ds.NearestDivisibleByM(rawDataOffset, 4) - rawDataOffset
	rawDataStripped := rawData
	if rawDataLength > alignedBytesCount {
		rawDataStripped = rawData[alignedBytesCount:]
	}

	return Inferences{
		IsObject:          meta2block.Inferences.IsObject,
		NumDirectChildren: meta2block.Inferences.NumDirectChildren,
		ParentIndex:       meta2block.Inferences.ParentIndex,
		HierarchyPath:     nil,
		RawDataOffset:     rawDataOffset,
		RawDataLength:     rawDataLength,
		RawDataStripped:   rawDataStripped,
	}
}

func InferHierarchyPath(index int, fields []Field) []string {
	// TODO: create a cache function if there is need for optimization
	fieldName := fields[index].Name
	parentIndex := fields[index].Inferences.ParentIndex
	if parentIndex == -1 {
		return []string{fields[index].Name}
	}
	return append(InferHierarchyPath(parentIndex, fields), fieldName)
}

func InferHierarchyPaths(fields []Field) []Field {
	fieldsCopy := lo.Map(
		fields,
		func(t Field, i int) Field {
			t.Inferences.HierarchyPath = InferHierarchyPath(i, fields)
			return t
		},
	)
	return fieldsCopy
}

func InferDataTypeByFieldName(fieldName string) DataType {
	switch fieldName {
	case "requirement_code":
		return DataTypeChar
	case
		"current_hp",
		"m_Stress":
		return DataTypeFloat
	case
		"read_page_indexes",
		"raid_page_indexes",
		"raid_unpage_indexes",
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
	case "killRange":
		return DataTypeTwoInt
	default:
		return DataTypeUnknown
	}
}

func InferDataTypeByHierarchyPath(hierarchyPath []string) DataType {
	_, resultAny := match.Match(hierarchyPath).
		When(
			// match.OneOf(
			// 	[]any{"actor", "buff", match.ANY, "amount"},
			// 	[]any{"chapter", match.ANY, match.ANY, "percent"},
			// ),
			[]any{"actor", "buff", match.ANY, "amount"},
			DataTypeFloat,
		).
		When(
			match.OneOf(
				[]any{"mash", "valid_additional_mash_entry_indexes"},
				[]any{"party", "heroes"},
				[]any{"curioGroups", match.ANY, "curios"},
				[]any{"curioGroups", match.ANY, "curio_table_entries"},
				[]any{"backer_heroes", match.ANY, "combat_skills"},
				[]any{"backer_heroes", match.ANY, "camping_skills"},
				[]any{"backer_heroes", match.ANY, "quirks"},
			),
			DataTypeIntVector,
		).
		When(
			match.OneOf(
				[]any{"roaming_dungeon_2_ids", match.ANY, "s"},
				[]any{"backgroundGroups", match.ANY, "backgrounds"},
				[]any{"backgroundGroups", match.ANY, "background_table_entries"},
			),
			DataTypeStringVector,
		).
		When(
			match.OneOf(
				[]any{"map", "bounds"},
				[]any{"areas", match.ANY, "bounds"},
				[]any{"areas", match.ANY, "tiles", match.ANY, "mappos"},
				[]any{"areas", match.ANY, "tiles", match.ANY, "sidepos"},
			),
			DataTypeFloatVector,
		).
		When(match.ANY, DataTypeUnknown).
		Result()
	result := resultAny.(DataType)
	return result
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
		b1 := rawDataStripped[3]
		b2 := rawDataStripped[7]
		oneOrZero := func(b byte) bool { return b == 0 || b == 1 }
		if oneOrZero(b1) && oneOrZero(b2) {
			return DataTypeTwoBool
		}
	case len(rawDataStripped) >= 8 &&
		bytes.Equal(rawDataStripped[4:8], dheader.MagicNumberBytes):
		return DataTypeFileRaw
	case len(rawDataStripped) >= 5:
		return DataTypeString
	}
	return DataTypeUnknown
}

func InferDataType(field Field) DataType {
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

func InferDataInt(rawDataStripped []byte) (int32, error) {
	if len(rawDataStripped) != 4 {
		err := fmt.Errorf(
			`InferDataInt got invalid input bytes length: expected 4; got "%v" with length %d`,
			rawDataStripped, len(rawDataStripped),
		)
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(rawDataStripped)), nil
}

func InferDataString(rawDataStripped []byte) (string, error) {
	rawLen := len(rawDataStripped)
	if len(rawDataStripped) < 4 {
		err := fmt.Errorf(
			`InferDataString got invalid input bytes length: expected value >= 4; got "%s" with length %d`,
			string(rawDataStripped), rawLen,
		)
		return "", err
	}

	strLen, err := InferDataInt(rawDataStripped[:4])
	if err != nil {
		err := errors.Wrapf(
			err, `InferDataString unreachable code with input bytes "%v"`,
			rawDataStripped,
		)
		return "", err
	}

	str := string(rawDataStripped[4:])
	trueLen := len(str)
	str = str[:len(str)-1]
	if trueLen != int(strLen) {
		err := fmt.Errorf(
			`InferDataString found unexpected string length for value "%s": expected %d; got %d`,
			str, strLen, trueLen,
		)
		return "", err
	}

	return str, nil
}

func InferDataChar(rawDataStripped []byte) (byte, error) {
	// TODO: check bytes length of all functions
	return rawDataStripped[0], nil
}

func InferDataBool(rawDataStripped []byte) (bool, error) {
	return rawDataStripped[0] == 1, nil
}

func InferDataFloat(rawDataStripped []byte) (float32, error) {
	return math.Float32frombits(
		binary.LittleEndian.Uint32(rawDataStripped),
	), nil
}

func InferDataIntVector(rawDataStripped []byte) ([]int32, error) {
	rawLen := len(rawDataStripped)
	if rawLen < 4 &&
		rawLen%4 != 0 {
		err := fmt.Errorf(
			`InferDataIntVector got invalid input bytes length: expected value >= 4 and divisible by 4; got "%v" with length %d`,
			rawDataStripped, rawLen,
		)
		return nil, err
	}

	intVectorLen, err := InferDataInt(rawDataStripped[:4])
	if err != nil {
		err := errors.Wrapf(
			err, `InferDataIntVector unreachable code with input bytes "%v"`,
			rawDataStripped,
		)
		return nil, err
	}

	intVectorByteChunks := ds.MakeChunks(rawDataStripped[4:], 4)
	if len(intVectorByteChunks) != int(intVectorLen) {
		err := fmt.Errorf(
			`InferDataIntVector got invalid input bytes length: expected %d; got "%v" with length %d`,
			intVectorLen, rawDataStripped, rawLen,
		)
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
		err := fmt.Errorf(
			`InferDataFloatVector got invalid input bytes length: expected value >= 4 and divisible by 4; got "%v" with length %d`,
			rawDataStripped, rawLen,
		)
		return nil, err
	}

	floatVectorLen, err := InferDataInt(rawDataStripped[:4])
	if err != nil {
		err := errors.Wrapf(
			err, `InferDataFloatVector unreachable code with input bytes "%v"`,
			rawDataStripped,
		)
		return nil, err
	}

	floatVectorByteChunks := ds.MakeChunks(rawDataStripped[4:], 4)
	if len(floatVectorByteChunks) != int(floatVectorLen) {
		err := fmt.Errorf(
			`InferDataFloatVector got invalid input bytes length: expected %d; got "%v" with length %d`,
			floatVectorLen, rawDataStripped, rawLen,
		)
		return nil, err
	}

	floatVector := make([]float32, 0, floatVectorLen)
	for _, byteChunk := range floatVectorByteChunks {
		value, err := InferDataFloat(byteChunk)
		if err != nil {
			err := fmt.Errorf(
				`InferDataFloatVector unreachable code with input bytes "%v"`,
				rawDataStripped,
			)
			return nil, err
		}
		floatVector = append(floatVector, value)
	}

	return floatVector, nil
}

func InferDataStringVector(rawDataStripped []byte) ([]string, error) {
	rawLen := len(rawDataStripped)
	if rawLen < 4 &&
		rawLen%4 != 0 {
		err := fmt.Errorf(
			`InferDataStringVector got invalid input bytes length: expected value >= 4 and divisible by 4; got "%v" with length %d`,
			rawDataStripped, rawLen,
		)
		return nil, err
	}

	stringVectorLen, err := InferDataInt(rawDataStripped[:4])
	if err != nil {
		err := fmt.Errorf(
			`InferDataStringVector unreachable code with input bytes "%v"`,
			rawDataStripped,
		)
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
			err := fmt.Errorf(
				`InferDataStringVector unreachable code with input bytes "%v"`,
				rawDataStripped,
			)
			return nil, err
		}
		cursor += 4 + len(str) + 1
		stringVector = append(stringVector, str)
	}

	return stringVector, nil
}

func InferDataTwoInt(rawDataStripped []byte) ([]int32, error) {
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
