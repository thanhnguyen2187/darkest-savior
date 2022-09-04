package dson

import (
	"encoding/json"
	"fmt"

	"darkest-savior/dson/dfield"
	"github.com/emirpasic/gods/maps/linkedhashmap"
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

func FromLinkedHashMap(lhm linkedhashmap.Map) (*DecodedFile, error) {
	file := DecodedFile{}
	revisionAny, ok := lhm.Get("__revision_dont_touch")
	if !ok {
		return nil, RevisionNotFoundError{}
	}
	file.Header.Revision = revisionAny.(int)
	lhm.Remove("__revision_dont_touch")

	for _, keyAny := range lhm.Keys() {
		keyStr, ok := keyAny.(string)
		if !ok {
			return nil, KeyCastError{Key: keyAny, Value: lhm.Get(keyAny)}
		}
		value, _ := lhm.Get(keyAny)
		field := dfield.Field{}
		field.Name = keyStr
		// field.Inferences.DataType = DecodeValue(value)
		field.Inferences.Data = value
	}

	return &file, nil
}

func EncodeDSON(jsonBytes []byte) ([]byte, error) {
	lhm := linkedhashmap.New()
	err := json.Unmarshal(jsonBytes, &lhm)
	if err != nil {
		return nil, err
	}

	_, err := FromLinkedHashMap(*lhm)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
