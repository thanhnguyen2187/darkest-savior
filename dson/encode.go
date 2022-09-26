package dson

import (
	"encoding/json"

	"darkest-savior/dson/dstruct"
	"github.com/iancoleman/orderedmap"
)

func EncodeJSON(fileBytes []byte) ([]byte, error) {
	// TODO: add `debug` parameter
	lhm := orderedmap.New()
	err := json.Unmarshal(fileBytes, lhm)
	if err != nil {
		return nil, err
	}
	dsonStruct, err := dstruct.FromLinkedHashMap(*lhm)
	if err != nil {
		return nil, err
	}
	resultBytes := dstruct.EncodeStruct(*dsonStruct)
	return resultBytes, nil
}
