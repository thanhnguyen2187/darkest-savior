package dson

import (
	"encoding/json"

	"github.com/iancoleman/orderedmap"
	"github.com/thanhnguyen2187/darkest-savior/dson/dstruct"
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
