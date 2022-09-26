package dfield

import (
	"fmt"

	"darkest-savior/dson/dhash"
	"darkest-savior/dson/dmeta2"
	"darkest-savior/dson/lbytes"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

func DecodeField(reader *lbytes.Reader, meta2Entry dmeta2.Entry) (*DataField, error) {
	// manual decoding and mapping is needed since turning data into JSON
	// and parse back does not work for bytes of RawData
	field := DataField{}
	err := error(nil)
	readString := lbytes.CreateStringReadFunction(reader, int(meta2Entry.Inferences.FieldNameLength))
	ok := false

	fieldName, err := readString()
	if err != nil {
		err := errors.Wrap(err, "DecodeField error: read field.Name")
		return nil, err
	}
	field.Name, ok = fieldName.(string)
	if !ok {
		err := fmt.Errorf(`DecodeField error: unable to cast value "%v" to string for field.Name`, fieldName)
		return nil, err
	}
	hashed := dhash.HashString(field.Name)
	if hashed != meta2Entry.NameHash {
		err := fmt.Errorf(
			`DecodeField error: mismatched hash value of field name "%s"; expected "%d", received "%d"`,
			field.Name, meta2Entry.NameHash, hashed,
		)
		return nil, err
	}

	field.RawData, err = reader.ReadBytes(int(meta2Entry.Inferences.RawDataLength))
	if err != nil {
		err := errors.Wrap(err, "DecodeField error: read field.RawData")
		return nil, err
	}

	field.Inferences = InferUsingMeta2Entry(field.RawData, meta2Entry)

	return &field, nil
}

func DecodeFields(reader *lbytes.Reader, meta2Blocks []dmeta2.Entry) ([]DataField, error) {
	fields := make([]DataField, 0, len(meta2Blocks))
	for _, meta2Block := range meta2Blocks {
		field, err := DecodeField(reader, meta2Block)
		if err != nil {
			err := errors.Wrap(err, "dfield.DecodeFields error")
			return nil, err
		}
		fields = append(fields, *field)
	}

	fields = InferHierarchyPaths(fields)
	fields = lo.Map(
		fields,
		func(field DataField, _ int) DataField {
			field.Inferences.DataType = InferDataType(field)
			return field
		},
	)
	for i, field := range fields {
		data, err := InferData(field.Inferences.DataType, field.Inferences.RawDataStripped)
		if err != nil {
			err := errors.Wrap(err, "dfield.DecodeFields error")
			return nil, err
		}
		field.Inferences.Data = data
		fields[i] = field
	}
	fields = lo.Map(
		fields,
		func(field DataField, _ int) DataField {
			return AttemptUnhashInt(field)
		},
	)
	fields = lo.Map(
		fields,
		func(field DataField, _ int) DataField {
			return AttemptUnhashIntVector(field)
		},
	)
	// TODO: investigate cases where the final vector includes both hashed and unhashed int.
	//       For example, `persists.tutorial.json` has all of `dispatched_events` converted,
	//       except `1972053455`.

	return fields, nil
}
