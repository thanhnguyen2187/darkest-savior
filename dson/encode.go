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

func FromLinkedHashMap(lhm orderedmap.OrderedMap) (*DecodedFile, error) {
	file := DecodedFile{}
	revisionAny, ok := lhm.Get("__revision_dont_touch")
	if !ok {
		return nil, RevisionNotFoundError{}
	}
	file.Header.Revision = revisionAny.(int)
	lhm.Delete("__revision_dont_touch")

	for _, key := range lhm.Keys() {
		if !ok {
			value, _ := lhm.Get(key)
			return nil, KeyCastError{Key: key, Value: value}
		}
		value, _ := lhm.Get(key)
		field := dfield.Field{}
		field.Name = key
		// field.Inferences.DataType = DecodeValue(value)
		field.Inferences.Data = value
	}

	return &file, nil
}

func EncodeDSON(jsonBytes []byte) ([]byte, error) {
	lhm := orderedmap.New()
	err := json.Unmarshal(jsonBytes, &lhm)
	if err != nil {
		return nil, err
	}

	// _, err = FromLinkedHashMap(*lhm)
	// if err != nil {
	// 	return nil, err
	// }

	lhm.Delete("__revision_dont_touch")
	value := dfield.FromLinkedHashMap(*lhm)
	print(ds.DumpJSON(value))

	return nil, nil
}
