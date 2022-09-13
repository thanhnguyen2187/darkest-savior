package dson

import (
	"encoding/json"
	"fmt"

	"darkest-savior/ds"
	"darkest-savior/dson/dfield"
	"github.com/iancoleman/orderedmap"
)

type (
	RevisionNotFoundError struct{}
	KeyCastError          struct {
		Key   any
		Value any
	}
)

func (RevisionNotFoundError) Error() string {
	return "expected __revision_dont_touch in input JSON file to convert to DSON"
}

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
	if err != nil {
		print(err.Error())
		return nil, err
	}
	return fields, nil
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
