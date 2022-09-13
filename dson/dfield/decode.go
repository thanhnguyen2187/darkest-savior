package dfield

import (
	"fmt"

	"darkest-savior/ds"
	"darkest-savior/dson/dhash"
	"darkest-savior/dson/dmeta2"
	"darkest-savior/dson/lbytes"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

func DecodeField(reader *lbytes.Reader, meta2Block dmeta2.Block) (*Field, error) {
	// manual decoding and mapping is needed since turning data into JSON
	// and parse back does not work for bytes of RawData
	field := Field{}
	err := error(nil)
	readString := lbytes.CreateStringReadFunction(reader, meta2Block.Inferences.FieldNameLength)
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
	if hashed != int32(meta2Block.NameHash) {
		err := fmt.Errorf(
			`DecodeField error: mismatched hash value of field name "%s"; expected "%d", received "%d"`,
			field.Name, meta2Block.NameHash, hashed,
		)
		return nil, err
	}

	field.RawData, err = reader.ReadBytes(meta2Block.Inferences.RawDataLength)
	if err != nil {
		err := errors.Wrap(err, "DecodeField error: read field.RawData")
		return nil, err
	}

	field.Inferences = InferUsingMeta2Block(field.RawData, meta2Block)

	return &field, nil
}

func DecodeFields(reader *lbytes.Reader, meta2Blocks []dmeta2.Block) ([]Field, error) {
	fields := make([]Field, 0, len(meta2Blocks))
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
		func(field Field, _ int) Field {
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
		func(field Field, _ int) Field {
			return AttemptUnhashInt(field)
		},
	)
	fields = lo.Map(
		fields,
		func(field Field, _ int) Field {
			return AttemptUnhashIntVector(field)
		},
	)
	// TODO: investigate cases where the final vector includes both hashed and unhashed int.
	//       For example, `persists.tutorial.json` has all of `dispatched_events` converted,
	//       except `1972053455`.

	return fields, nil
}

func RemoveDuplications(fields []Field) []Field {
	type Tracker struct {
		NumRemainChildren int
		NumAllChildren    int
		EncounteredNames  map[string]struct{}
	}
	stack := ds.NewStack[Tracker]()
	stack.Push(
		Tracker{
			NumRemainChildren: 1,
			NumAllChildren:    1,
			EncounteredNames:  map[string]struct{}{},
		},
	)

	uniqueFields := make([]Field, 0, len(fields))
	skipping := 0
	for _, field := range fields {
		if skipping > 0 {
			skipping -= 1
			continue
		}
		stack.ReplaceLast(
			func(tr Tracker) Tracker {
				tr.NumRemainChildren -= 1
				return tr
			},
		)
		last := stack.Peek()

		_, ok := last.EncounteredNames[field.Name]
		if !ok {
			uniqueFields = append(uniqueFields, field)
		} else {
			skipping += field.Inferences.NumAllChildren
			// continue
		}
		last.EncounteredNames[field.Name] = struct{}{}

		if last.NumRemainChildren == 0 {
			stack.Pop()
		}
		if field.Inferences.IsObject {
			stack.Push(
				Tracker{
					NumRemainChildren: field.Inferences.NumDirectChildren,
					EncounteredNames:  map[string]struct{}{},
				},
			)
		}
	}

	return uniqueFields
}
