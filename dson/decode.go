package dson

import (
	"encoding/json"

	"darkest-savior/dson/dfield"
	"darkest-savior/dson/dheader"
	"darkest-savior/dson/dmeta1"
	"darkest-savior/dson/dmeta2"
	"darkest-savior/dson/lbytes"
	"github.com/iancoleman/orderedmap"
	"github.com/samber/lo"
)

type (
	DecodedFile struct {
		Header     dheader.Header `json:"header"`
		Meta1Block []dmeta1.Entry `json:"meta_1_block"`
		Meta2Block []dmeta2.Entry `json:"meta_2_block"`
		Fields     []dfield.Field `json:"fields"`
	}
)

func ToStructuredFile(bs []byte) (*DecodedFile, error) {
	reader := lbytes.NewBytesReader(bs)
	file := DecodedFile{}
	err := error(nil)

	header, err := dheader.Decode(reader)
	if err != nil {
		return nil, err
	}
	file.Header = *header
	file.Meta1Block, err = dmeta1.DecodeBlock(reader, header.NumMeta1Entries)
	if err != nil {
		return nil, err
	}

	file.Meta2Block, err = dmeta2.DecodeBlock(reader, file.Header, file.Meta1Block)
	if err != nil {
		return nil, err
	}

	file.Fields, err = dfield.DecodeFields(reader, file.Meta2Block)
	if err != nil {
		return nil, err
	}
	// file.Fields = dfield.RemoveDuplications(file.Fields)

	for i := range file.Fields {
		field := &file.Fields[i]
		// handle case there is an embedded file
		// ideally, the code should be put into `dfield`,
		// but it would create a circular dependency between the package and `dson`
		if field.Inferences.DataType == dfield.DataTypeFileRaw {
			rawDataSkipped := field.Inferences.RawDataStripped[4:]
			embeddedFile, err := ToStructuredFile(rawDataSkipped)
			if err != nil {
				return nil, err
			}
			field.Inferences.Data = *embeddedFile
			field.Inferences.DataType = dfield.DataTypeFileDecoded
		}
	}

	return &file, nil
}

func DecodeDSON(bytes []byte, debug bool) ([]byte, error) {
	decodedFile, err := ToStructuredFile(bytes)
	if err != nil {
		return nil, err
	}

	if debug {
		decodedFileBytes := make([]byte, 0)
		decodedFileBytes, err = json.MarshalIndent(decodedFile, "", "  ")
		return decodedFileBytes, nil
	}

	decodedMap := ToLinkedHashMap(*decodedFile)
	decodedBytes, err := json.MarshalIndent(decodedMap, "", "  ")
	return decodedBytes, nil
}

func ToLinkedHashMap(file DecodedFile) *orderedmap.OrderedMap {
	// TODO: use an interface for orderedMap
	lhmByIndex := make(map[int]*orderedmap.OrderedMap)
	lhmByIndex[-1] = orderedmap.New()
	lhmByIndex[-1].Set(dfield.FieldNameRevision, file.Header.Revision)
	for index, field := range file.Fields {
		parentIndex := field.Inferences.ParentIndex
		if field.Inferences.IsObject {
			lhm := orderedmap.New()
			lhmByIndex[index] = lhm
			lhmByIndex[parentIndex].Set(field.Name, lhm)
		} else if field.Inferences.DataType == dfield.DataTypeFileDecoded {
			lhm := ToLinkedHashMap(field.Inferences.Data.(DecodedFile))
			lhmByIndex[parentIndex].Set(field.Name, lhm)
		} else {
			lhmByIndex[parentIndex].Set(field.Name, field.Inferences.Data)
		}
	}

	return lhmByIndex[-1]
}

func ExpandEmbeddedFiles(fields []dfield.Field) ([]dfield.Field, error) {
	mappedFields := make([][]dfield.Field, 0)
	for _, field := range fields {
		mappedFields = append(mappedFields, []dfield.Field{field})
		if field.Inferences.DataType == dfield.DataTypeFileDecoded {
			decodedFileBytes, err := json.Marshal(field.Inferences.Data)
			if err != nil {
				return nil, err
			}
			decodedFile := DecodedFile{}
			err = json.Unmarshal(decodedFileBytes, &decodedFile)
			if err != nil {
				return nil, err
			}
			mappedFields = append(mappedFields, decodedFile.Fields)
		}
	}
	return lo.Flatten(mappedFields), nil
}
