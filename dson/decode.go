package dson

import (
	"darkest-savior/ds"
	"darkest-savior/dson/dfield"
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

	return &file, nil
}

func ToLinkedHashMap(fields []dfield.Field) ds.LinkedHashMap[any, any] {
	lhMapByIndex := map[int]*ds.LinkedHashMap[any, any]{}
	lhMapByIndex[-1] = ds.NewLinkedHashMap[any, any]()
	for i, field := range fields {
		parentIndex := field.Inferences.ParentIndex
		lhMapParent := lhMapByIndex[parentIndex]
		if field.Inferences.IsObject {
			lhMap := ds.NewLinkedHashMap[any, any]()
			lhMapByIndex[i] = lhMap
			lhMapParent.Put(field.Name, lhMap)
		} else {
			lhMapParent.Put(field.Name, field.Inferences.Data)
		}
	}
	return *lhMapByIndex[-1]
}
