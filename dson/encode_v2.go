package dson

import (
	"darkest-savior/ds"
	"darkest-savior/dson/dfield"
	"darkest-savior/dson/dhash"
	"darkest-savior/dson/dheader"
	"darkest-savior/dson/dmeta1"
	"darkest-savior/dson/dmeta2"
	"darkest-savior/dson/lbytes"
	"github.com/iancoleman/orderedmap"
	"github.com/samber/lo"
)

func LHMToDataFields(parentHierarchyPath []string, lhm orderedmap.OrderedMap) ([]dfield.DataField, error) {
	dataFields := make([]dfield.DataField, 0, len(lhm.Keys()))
	for _, key := range lhm.Keys() {
		if key == dfield.FieldNameRevision {
			continue
		}
		value, _ := lhm.Get(key)
		fieldName := key
		hierarchyPath := ds.ShallowCopy(parentHierarchyPath)
		hierarchyPath = append(hierarchyPath, fieldName)
		field := dfield.DataField{}

		field.Name = fieldName
		field.Inferences.HierarchyPath = hierarchyPath
		dataType := dfield.ImplyDataType(key, hierarchyPath[1:], value)
		switch dataType {
		case dfield.DataTypeObject:
			valueLhm := value.(orderedmap.OrderedMap)
			childFields, err := LHMToDataFields(hierarchyPath, valueLhm)
			if err != nil {
				return nil, err
			}
			field.Inferences.IsObject = true
			field.Inferences.Data = nil
			field.Inferences.RawDataStripped = make([]byte, 0)
			dataFields = append(dataFields, field)
			dataFields = append(dataFields, childFields...)
		case dfield.DataTypeFileJSON:
			valueLhm := value.(orderedmap.OrderedMap)
			embeddedStruct, err := FromLinkedHashMapV2(valueLhm)
			if err != nil {
				return nil, err
			}
			embeddedStructBytes := EncodeStruct(*embeddedStruct)
			field.Inferences.RawDataStripped = append(
				lbytes.EncodeValueInt(len(embeddedStructBytes)),
				embeddedStructBytes...,
			)
			field.Inferences.Data = *embeddedStruct
			field.Inferences.DataType = dfield.DataTypeFileDecoded
			dataFields = append(dataFields, field)
		default:
			rawDataStripped, err := dfield.EncodeValue(key, dataType, value)
			if err != nil {
				return nil, err
			}
			field.Inferences.Data = value
			field.Inferences.RawDataStripped = rawDataStripped
			dataFields = append(dataFields, field)
		}
	}
	return dataFields, nil
}

func FromLinkedHashMapV2(lhm orderedmap.OrderedMap) (*Struct, error) {
	lhm = ds.Deref(&lhm)
	revisionKey := lhm.Keys()[0]
	if revisionKey != dfield.FieldNameRevision {
		return nil, dfield.RevisionNotFoundError{
			ActualFieldName: revisionKey,
		}
	}
	revisionAny, _ := lhm.Get(revisionKey)
	revision := int32(revisionAny.(float64))

	dataFields, err := LHMToDataFields([]string{}, lhm)
	if err != nil {
		return nil, err
	}
	numsAllChildren := dfield.CalculateNumAllChildren(lhm)
	// drop one key of revision
	numsAllChildren = lo.Drop(numsAllChildren, 1)
	parentIndexes := dfield.CalculateParentIndexes(numsAllChildren)
	numsDirectChildren := dfield.CalculateNumDirectChildren(lhm)
	numsDirectChildren = lo.Drop(numsDirectChildren, 1)

	fieldNameLengths := lo.Map(
		dataFields,
		func(t dfield.DataField, _ int) int { return len(t.Name) + 1 },
	)
	rawDataStrippedLengths := lo.Map(
		dataFields,
		func(t dfield.DataField, _ int) int { return len(t.Inferences.RawDataStripped) },
	)
	isObjectSlice := lo.Map(
		dataFields,
		func(t dfield.DataField, _ int) bool { return t.Inferences.IsObject },
	)
	meta2Offsets := dfield.CalculateMeta2OffsetsV2(
		fieldNameLengths,
		rawDataStrippedLengths,
	)
	meta2OffsetsDropped := lo.DropRight(meta2Offsets, 1)
	paddedBytesCounts := dfield.CalculatePaddedBytesCountsV2(
		fieldNameLengths,
		meta2OffsetsDropped,
		rawDataStrippedLengths,
	)
	meta1EntryIndexes := dfield.CalculateMeta1EntryIndexesV2(isObjectSlice)

	dataFields = lo.Map(
		lo.Zip3(dataFields, parentIndexes, paddedBytesCounts),
		func(tuple lo.Tuple3[dfield.DataField, int32, int], _ int) dfield.DataField {
			field := tuple.A
			parentIndex := tuple.B
			paddedBytesCount := tuple.C
			field.Inferences.ParentIndex = parentIndex
			field.RawData = append(
				lbytes.CreateZeroBytes(paddedBytesCount),
				field.Inferences.RawDataStripped...,
			)
			return field
		},
	)

	fieldInfoSlice := lo.Map(
		lo.Zip3(fieldNameLengths, isObjectSlice, meta1EntryIndexes),
		func(tuple lo.Tuple3[int, bool, int32], _ int) int32 {
			fieldNameLength := tuple.A
			isObject := tuple.B
			meta1EntryIndex := tuple.C
			fieldInfo := dmeta2.CalculateFieldInfo(fieldNameLength, isObject, meta1EntryIndex)
			return fieldInfo
		},
	)

	meta2Block := lo.Map(
		lo.Zip3(dataFields, meta2OffsetsDropped, fieldInfoSlice),
		func(tuple lo.Tuple3[dfield.DataField, int32, int32], _ int) dmeta2.Entry {
			field := tuple.A
			meta2Offset := tuple.B
			fieldInfo := tuple.C
			entry := dmeta2.Entry{
				NameHash:  dhash.HashString(field.Name),
				Offset:    meta2Offset,
				FieldInfo: fieldInfo,
			}
			return entry
		},
	)

	meta1Block := lo.FilterMap(
		lo.Zip3(numsDirectChildren, numsAllChildren, isObjectSlice),
		func(tuple lo.Tuple3[int32, int32, bool], index int) (dmeta1.Entry, bool) {
			numDirectChildren := tuple.A
			numAllChildren := tuple.B
			isObject := tuple.C

			if !isObject {
				return dmeta1.Entry{}, false
			}

			parentIndex := parentIndexes[index]
			meta1ParentIndex := int32(0)
			if parentIndex != -1 {
				meta1ParentIndex = meta1EntryIndexes[parentIndex]
			} else {
				meta1ParentIndex = -1
			}
			meta2EntryIndex := index
			return dmeta1.Entry{
				ParentIndex:       meta1ParentIndex,
				Meta2EntryIndex:   int32(meta2EntryIndex),
				NumDirectChildren: numDirectChildren,
				NumAllChildren:    numAllChildren,
			}, true
		},
	)

	headerLength := int32(dheader.DefaultHeaderSize)
	numMeta1Entries := int32(len(meta1Block))
	numMeta2Entries := int32(len(meta2Block))
	meta1FirstOffset := headerLength
	meta1Size := numMeta1Entries * dmeta1.DefaultEntrySize
	meta2FirstOffset := meta1FirstOffset + meta1Size
	meta2Size := numMeta2Entries * dmeta2.DefaultEntrySize
	dataOffset := headerLength + meta1Size + meta2Size
	dataLength, _ := lo.Last(meta2Offsets)
	header := dheader.Header{
		MagicNumber:     dheader.MagicNumberBytes,
		Revision:        revision,
		HeaderLength:    dheader.DefaultHeaderSize,
		Zeroes:          lbytes.CreateZeroBytes(4),
		Meta1Size:       meta1Size,
		NumMeta1Entries: numMeta1Entries,
		Meta1Offset:     meta1FirstOffset,
		Zeroes2:         lbytes.CreateZeroBytes(8),
		Zeroes3:         lbytes.CreateZeroBytes(8),
		NumMeta2Entries: numMeta2Entries,
		Meta2Offset:     meta2FirstOffset,
		Zeroes4:         lbytes.CreateZeroBytes(4),
		DataLength:      dataLength,
		DataOffset:      dataOffset,
	}

	decodedFile := Struct{
		Header:     header,
		Meta1Block: meta1Block,
		Meta2Block: meta2Block,
		Fields:     dataFields,
	}
	return &decodedFile, nil
}
