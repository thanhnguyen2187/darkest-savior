package dson

import (
	"encoding/json"
	"fmt"

	"darkest-savior/ds"
	"darkest-savior/dson/dfield"
	"github.com/iancoleman/orderedmap"
)

type (
	KeyCastError struct {
		Key   any
		Value any
	}
)

func (r KeyCastError) Error() string {
	return fmt.Sprintf(
		`unable to cast key "%v" of value "%v" to string`,
		r.Key, r.Value,
	)
}

func FromLinkedHashMap(lhm orderedmap.OrderedMap) ([]dfield.EncodingField, error) {
	lhm = ds.Deref(&lhm)
	fields := dfield.FromLinkedHashMap([]string{}, lhm)
	fields, err := dfield.EncodeValues(fields)
	fields = dfield.SetIndexes(fields)
	numsDirectChildren := dfield.CalculateNumDirectChildren(lhm)
	fields = dfield.SetNumDirectChildren(fields, numsDirectChildren)
	numsAllChildren := dfield.CalculateNumAllChildren(lhm)
	fields = dfield.SetNumAllChildren(fields, numsAllChildren)
	parentIndexes := dfield.CalculateParentIndexes(numsAllChildren)
	fields = dfield.SetParentIndexes(fields, parentIndexes)
	meta1EntryIndexes := dfield.CalculateMeta1EntryIndexes(fields)
	fields = dfield.SetMeta1EntryIndexes(fields, meta1EntryIndexes)
	fields = dfield.SetMeta1ParentIndexes(fields)
	if err != nil {
		print(err.Error())
		return nil, err
	}
	return fields, nil
}

func CompactEmbeddedFiles(fields []dfield.EncodingField) []dfield.EncodingField {
	skipping := 0
	for _, field := range fields {
		if skipping > 0 {
			skipping -= 1
			continue
		}
		if field.ValueType == dfield.DataTypeFileJSON {
			// startIndex := index + 1
			// endIndex := startIndex + field.NumAllChildren
			// skipping += field.NumAllChildren
			// embeddedFileFields := fields[startIndex:endIndex]

			// header, err := dfield.CreateHeader(embeddedFileFields)
			// meta1Block, err := dfield.CreateMeta1Block(embeddedFileFields)
			// meta2Block, err := dfield.CreateMeta2Block(embeddedFileFields)
		}
	}
	return nil
}

func EncodeDSON(jsonBytes []byte) ([]byte, error) {
	lhm := orderedmap.New()
	err := json.Unmarshal(jsonBytes, &lhm)
	if err != nil {
		return nil, err
	}

	fields, err := FromLinkedHashMap(*lhm)
	if err != nil {
		return nil, err
	}
	fieldsBytes, err := json.MarshalIndent(fields, "", "  ")
	if err != nil {
		return nil, err
	}
	// print(ds.DumpJSON(fields))

	return fieldsBytes, nil
}
