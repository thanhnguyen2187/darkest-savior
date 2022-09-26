package dson

import (
	"encoding/json"

	"darkest-savior/dson/dstruct"
)

func DecodeDSON(bytes []byte, debug bool) ([]byte, error) {
	decodedFile, err := dstruct.ToStructuredFile(bytes)
	if err != nil {
		return nil, err
	}

	if debug {
		decodedFileBytes := make([]byte, 0)
		decodedFileBytes, err = json.MarshalIndent(decodedFile, "", "  ")
		return decodedFileBytes, nil
	}

	decodedMap := dstruct.ToLinkedHashMap(*decodedFile)
	decodedBytes, err := json.MarshalIndent(decodedMap, "", "  ")
	return decodedBytes, nil
}
