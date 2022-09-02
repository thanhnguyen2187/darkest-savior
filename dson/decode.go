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

func DecodeDSON(reader *lbytes.Reader) (*DecodedFile, error) {
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
		if field.Inferences.DataType == dfield.DataTypeFileRaw {
			// embedded files have their first 4 bytes denote the total length
			reader := lbytes.NewBytesReader(field.Inferences.RawDataStripped[4:])
			embeddedFile, err := DecodeDSON(reader)
			if err != nil {
				return nil, err
			}
			field.Inferences.Data = *embeddedFile
			field.Inferences.DataType = dfield.DataTypeFileDecoded
		}
		if field.Inferences.DataType == dfield.DataTypeInt {
			name, ok := dhash.NameByHash[field.Inferences.Data.(int32)]
			if ok {
				field.Inferences.Data = name
				field.Inferences.DataType = dfield.DataTypeHashedInt
			}
		}
	}
	return &file, nil
}

func ToLinkedHashMap(fields []dfield.Field) ds.LinkedHashMap[any, any] {
	lhmByIndex := make(map[int]*ds.LinkedHashMap[any, any])
	lhmByIndex[-1] = ds.NewLinkedHashMap[any, any]()
	for index, field := range fields {
		lhm := &ds.LinkedHashMap[any, any]{}
		parentIndex := field.Inferences.ParentIndex
		if field.Inferences.IsObject {
			lhm = ds.NewLinkedHashMap[any, any]()
			lhmByIndex[index] = lhm
			lhmByIndex[parentIndex].Put(field.Name, lhm)
		} else {
			lhmByIndex[parentIndex].Put(field.Name, field.Inferences.Data)
		}
	}

	return *lhmByIndex[-1]
}
