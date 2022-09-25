package dson

import (
	"encoding/json"
	"fmt"

	"darkest-savior/ds"
	"darkest-savior/dson/dfield"
	"darkest-savior/dson/dheader"
	"darkest-savior/dson/dmeta1"
	"darkest-savior/dson/dmeta2"
	"darkest-savior/dson/lbytes"
	"github.com/iancoleman/orderedmap"
	"github.com/samber/lo"
)

type (
	KeyCastError struct {
		Key   any
		Value any
	}
)

func (r KeyCastError) Error() string {
	return fmt.Sprintf(
		`unable to cast key "%v" of value "%v" to string`,
		r.Key, r.Value,
	)
}

func FromLinkedHashMap(lhm orderedmap.OrderedMap) (*dfield.EncodingFieldsWithRevision, error) {
	lhm = ds.Deref(&lhm)
	fields := dfield.FromLinkedHashMap([]string{}, lhm)
	revision, err := dfield.ExtractRevision(fields)
	if err != nil {
		return nil, err
	}
	fields = dfield.RemoveRevisionField(fields)
	lhm.Delete(dfield.FieldNameRevision)
	{
		fields, err = dfield.EncodeValues(fields)
		if err != nil {
			return nil, err
		}
		fields = dfield.SetIndexes(fields)
	}
	{
		numsDirectChildren := dfield.CalculateNumDirectChildren(lhm)
		fields = dfield.SetNumDirectChildren(fields, numsDirectChildren)
	}
	{
		numsAllChildren := dfield.CalculateNumAllChildren(lhm)
		fields = dfield.SetNumAllChildren(fields, numsAllChildren)
		parentIndexes := dfield.CalculateParentIndexes(numsAllChildren)
		fields = dfield.SetParentIndexes(fields, parentIndexes)
	}
	{
		meta1EntryIndexes := dfield.CalculateMeta1EntryIndexes(fields)
		fields = dfield.SetMeta1EntryIndexes(fields, meta1EntryIndexes)
	}
	{
		meta2Offsets := dfield.CalculateMeta2Offsets(fields)
		fields = dfield.SetMeta2Offsets(fields, lo.DropRight(meta2Offsets, 1))
	}
	{
		paddedBytesCounts := dfield.CalculatePaddedBytesCounts(fields)
		fields = dfield.SetPaddedBytesCounts(fields, paddedBytesCounts)
	}
	{
		fields = dfield.SetMeta1ParentIndexes(fields)
	}
	fieldsWithRevision := dfield.EncodingFieldsWithRevision{
		Revision: revision,
		Fields:   fields,
	}
	return &fieldsWithRevision, nil
}

func CompactEmbeddedFiles(fields []dfield.EncodingField) ([]dfield.EncodingField, error) {
	resultFields := make([]dfield.EncodingField, 0, len(fields))
	skipping := int32(0)
	for index, field := range fields {
		if skipping > 0 {
			skipping -= 1
			continue
		}
		if field.ValueType == dfield.DataTypeFileJSON {
			startIndex := int32(index) + 1
			endIndex := startIndex + field.NumAllChildren
			skipping += field.NumAllChildren
			embeddedFileFields := fields[startIndex:endIndex]

			decodedFile, err := CreateDecodedFile(embeddedFileFields)
			if err != nil {
				return nil, err
			}
			embeddedFileBytes := EncodeStruct(*decodedFile)
			embeddedFileBytes = append(
				embeddedFileBytes,
				lbytes.EncodeValueInt(len(embeddedFileBytes))...,
			)

			resultField := dfield.EncodingField{
				Index:             field.Index,
				Key:               field.Key,
				ValueType:         dfield.DataTypeFileDecoded,
				Value:             *decodedFile,
				Bytes:             embeddedFileBytes,
				PaddedBytesCount:  0,
				IsObject:          false,
				ParentIndex:       field.ParentIndex,
				Meta1ParentIndex:  field.Meta1ParentIndex,
				Meta1EntryIndex:   field.Meta1EntryIndex,
				Meta2Offset:       field.Meta2Offset,
				NumDirectChildren: 0,
				NumAllChildren:    0,
				HierarchyPath:     field.HierarchyPath,
			}
			resultFields = append(resultFields, resultField)
		} else {
			resultFields = append(resultFields, field)
		}
	}
	return resultFields, nil
}

func EncodeDSON(jsonBytes []byte) ([]byte, error) {
	lhm := orderedmap.New()
	err := json.Unmarshal(jsonBytes, &lhm)
	if err != nil {
		return nil, err
	}

	fields, err := FromLinkedHashMap(*lhm)
	if err != nil {
		return nil, err
	}
	fieldsBytes, err := json.MarshalIndent(fields, "", "  ")
	if err != nil {
		return nil, err
	}
	// print(ds.DumpJSON(fields))

	return fieldsBytes, nil
}

func PickBytesHeader(fileBytes []byte) []byte {
	return fileBytes[:dheader.DefaultHeaderSize]
}

func PickBytesMeta1Block(numMeta1Entries int, fileBytes []byte) []byte {
	// change PickBytes to use more data from Header
	start := dheader.DefaultHeaderSize
	blockLength := dmeta1.CalculateBlockLength(numMeta1Entries)
	end := start + blockLength
	return fileBytes[start:end]
}

func PickBytesMeta2Block(numMeta1Entries int, numMeta2Entries int, fileBytes []byte) []byte {
	meta1Offset := dheader.DefaultHeaderSize
	meta2Offset := meta1Offset + dmeta1.CalculateBlockLength(numMeta1Entries)
	meta2Length := meta2Offset + numMeta2Entries*dmeta2.DefaultEntrySize

	return fileBytes[meta2Offset:meta2Length]
}

func EncodeStruct(file Struct) []byte {
	headerSize := dheader.DefaultHeaderSize
	meta1Size := int(file.Header.Meta1Size)
	meta2Size := dmeta2.CalculateBlockSize(int(file.Header.NumMeta2Entries))
	dataSize := int(file.Header.DataLength)
	totalSize := headerSize + meta1Size + meta2Size + dataSize

	headerBytes := dheader.Encode(file.Header)
	meta1Bytes := dmeta1.EncodeBlock(file.Meta1Block)
	meta2Bytes := dmeta2.EncodeBlock(file.Meta2Block)
	dataFieldsBytes := dfield.EncodeDataFields(file.Fields)

	bs := make([]byte, 0, totalSize)
	bs = append(bs, headerBytes...)
	bs = append(bs, meta1Bytes...)
	bs = append(bs, meta2Bytes...)
	bs = append(bs, dataFieldsBytes...)
	return bs
}

func CreateDecodedFile(fields []dfield.EncodingField) (*Struct, error) {
	header, err := dfield.CreateHeader(fields)
	if err != nil {
		return nil, err
	}
	meta1Block := dfield.CreateMeta1Block(fields[1:])
	meta2Block := dfield.CreateMeta2Block(fields[1:])
	dataFields := dfield.CreateDataFields(fields[1:])
	return &Struct{
		Header:     *header,
		Meta1Block: meta1Block,
		Meta2Block: meta2Block,
		Fields:     dataFields,
	}, nil
}
