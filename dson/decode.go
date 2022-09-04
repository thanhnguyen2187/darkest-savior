package dson

import (
	"darkest-savior/dson/dfield"
	"darkest-savior/dson/dheader"
	"darkest-savior/dson/dmeta1"
	"darkest-savior/dson/dmeta2"
	"darkest-savior/dson/lbytes"
	"github.com/emirpasic/gods/maps/linkedhashmap"
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
		// handle case there is an embedded file
		// ideally, the code should be put into `dfield`,
		// but it would create a circular dependency between the package and `dson`
		if field.Inferences.DataType == dfield.DataTypeFileRaw {
			rawDataSkipped := field.Inferences.RawDataStripped[4:]
			embeddedFile, err := DecodeDSON(rawDataSkipped)
			if err != nil {
				return nil, err
			}
			field.Inferences.Data = *embeddedFile
			field.Inferences.DataType = dfield.DataTypeFileDecoded
		}
	}
	return &file, nil
}

func ToLinkedHashMap(file DecodedFile) *linkedhashmap.Map {
	lhmByIndex := make(map[int]*linkedhashmap.Map)
	lhmByIndex[-1] = linkedhashmap.New()
	lhmByIndex[-1].Put("__revision_dont_touch", file.Header.Revision)
	for index, field := range file.Fields {
		parentIndex := field.Inferences.ParentIndex
		if field.Inferences.IsObject {
			lhm := linkedhashmap.New()
			lhmByIndex[index] = lhm
			lhmByIndex[parentIndex].Put(field.Name, lhm)
		} else if field.Inferences.DataType == dfield.DataTypeFileDecoded {
			lhm := ToLinkedHashMap(field.Inferences.Data.(DecodedFile))
			lhmByIndex[parentIndex].Put(field.Name, lhm)
		} else {
			lhmByIndex[parentIndex].Put(field.Name, field.Inferences.Data)
		}
	}

	return lhmByIndex[-1]
}
