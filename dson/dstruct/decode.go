package dstruct

import (
	"darkest-savior/dson/dfield"
	"darkest-savior/dson/dheader"
	"darkest-savior/dson/dmeta1"
	"darkest-savior/dson/dmeta2"
	"darkest-savior/dson/lbytes"
	"github.com/iancoleman/orderedmap"
)

func ToLinkedHashMap(file Struct) *orderedmap.OrderedMap {
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
			lhm := ToLinkedHashMap(field.Inferences.Data.(Struct))
			lhmByIndex[parentIndex].Set(field.Name, lhm)
		} else {
			lhmByIndex[parentIndex].Set(field.Name, field.Inferences.Data)
		}
	}

	return lhmByIndex[-1]
}

func ToStructuredFile(bs []byte) (*Struct, error) {
	reader := lbytes.NewBytesReader(bs)
	file := Struct{}
	err := error(nil)

	header, err := dheader.Decode(reader)
	if err != nil {
		return nil, err
	}
	file.Header = *header
	file.Meta1Block, err = dmeta1.DecodeBlock(reader, int(header.NumMeta1Entries))
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
