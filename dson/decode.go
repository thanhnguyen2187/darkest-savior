package dson

import (
	"darkest-savior/ds"
	"darkest-savior/dson/dfield"
	"darkest-savior/dson/dhash"
	"darkest-savior/dson/dheader"
	"darkest-savior/dson/dmeta1"
	"darkest-savior/dson/dmeta2"
	"darkest-savior/dson/lbytes"
)

type (
	DecodedFile struct {
		Header      dheader.Header `json:"header"`
		Meta1Blocks []dmeta1.Block `json:"meta_1_blocks"`
		Meta2Blocks []dmeta2.Block `json:"meta_2_blocks"`
		Fields      []dfield.Field `json:"fields"`
	}
)

func DecodeDSON(bytes []byte) (*DecodedFile, error) {
	reader := lbytes.NewBytesReader(bytes)
	file := DecodedFile{}
	err := error(nil)

	header, err := dheader.Decode(reader)
	if err != nil {
		return nil, err
	}
	file.Header = *header
	file.Meta1Blocks, err = dmeta1.DecodeBlocks(reader, header.NumMeta1Entries)
	if err != nil {
		return nil, err
	}

	file.Meta2Blocks, err = dmeta2.DecodeBlocks(reader, file.Header, file.Meta1Blocks)
	if err != nil {
		return nil, err
	}

	file.Fields, err = dfield.DecodeFields(reader, file.Meta2Blocks)
	if err != nil {
		return nil, err
	}

	for i := range file.Fields {
		field := &file.Fields[i]
		switch field.Inferences.DataType {
		case dfield.DataTypeFileRaw:
			rawDataSkipped := field.Inferences.RawDataStripped[4:]
			embeddedFile, err := DecodeDSON(rawDataSkipped)
			if err != nil {
				return nil, err
			}
			field.Inferences.Data = *embeddedFile
			field.Inferences.DataType = dfield.DataTypeFileDecoded
		case dfield.DataTypeInt:
			name, ok := dhash.NameByHash[field.Inferences.Data.(int32)]
			if ok {
				field.Inferences.Data = name
				field.Inferences.DataType = dfield.DataTypeHashedInt
			}
		case dfield.DataTypeIntVector:
			hashedNames := field.Inferences.Data.([]int32)
			converted := false
			names := make([]any, 0, len(hashedNames))
			for _, hashedName := range hashedNames {
				if name, ok := dhash.NameByHash[hashedName]; ok {
					converted = true
					names = append(names, name)
				} else {
					names = append(names, hashedName)
				}
			}
			// TODO: investigate cases where the final vector includes both hashed and unhashed int.
			//       For example, `persists.tutorial.json` has all of `dispatched_events` converted,
			//       except `1972053455`.
			if converted {
				field.Inferences.Data = names
				field.Inferences.DataType = dfield.DataTypeHashedIntVector
			}
		}
	}
	return &file, nil
}

func ToLinkedHashMap(file DecodedFile) ds.LinkedHashMap[any, any] {
	lhmByIndex := make(map[int]*ds.LinkedHashMap[any, any])
	lhmByIndex[-1] = ds.NewLinkedHashMap[any, any]()
	lhmByIndex[-1].Put("__revision_dont_touch", file.Header.Revision)
	for index, field := range file.Fields {
		lhm := &ds.LinkedHashMap[any, any]{}
		parentIndex := field.Inferences.ParentIndex
		if field.Inferences.IsObject {
			lhm = ds.NewLinkedHashMap[any, any]()
			lhmByIndex[index] = lhm
			lhmByIndex[parentIndex].Put(field.Name, lhm)
		} else if field.Inferences.DataType == dfield.DataTypeFileDecoded {
			lhm2 := ToLinkedHashMap(field.Inferences.Data.(DecodedFile))
			lhm = &lhm2
			lhmByIndex[parentIndex].Put(field.Name, lhm)
		} else {
			lhmByIndex[parentIndex].Put(field.Name, field.Inferences.Data)
		}
	}

	return *lhmByIndex[-1]
}
