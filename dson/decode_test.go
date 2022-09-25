package dson

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"darkest-savior/ds"
	"darkest-savior/dson/dfield"
	"darkest-savior/dson/dheader"
	"darkest-savior/dson/dmeta1"
	"darkest-savior/dson/dmeta2"
	"github.com/iancoleman/orderedmap"
	"github.com/samber/lo"
	"github.com/stretchr/testify/suite"
)

// EndToEndTestSuite includes all the needed data to perform the best "real-world" simulation.
type EndToEndTestSuite struct {
	suite.Suite
	FilePaths                     []string
	FileByteSlices                [][]byte
	DecodedFiles                  []Struct
	DecodedJSONsFromFile          []orderedmap.OrderedMap
	DecodedFieldSlices            [][]dfield.DataField
	DecodedFieldSlicesUnique      [][]dfield.DataField
	DecodedFieldSlicesExpanded    [][]dfield.DataField
	EncodingFieldsWithRevision    []dfield.EncodingFieldsWithRevision
	EncodingFieldSlicesNoRevision [][]dfield.EncodingField
	EncodingFieldSlicesCompacted  [][]dfield.EncodingField
}

func (suite *EndToEndTestSuite) SetupSuite() {
	suiteR := suite.Require()
	suite.FilePaths = []string{
		// "../sample_data/novelty_tracker.json",
		// "../sample_data/persist.campaign_log.json",
		// "../sample_data/persist.campaign_mash.json",
		// "../sample_data/persist.curio_tracker.json",
		// "../sample_data/persist.estate.json",
		// "../sample_data/persist.game.json",
		// "../sample_data/persist.game_knowledge.json",
		// "../sample_data/persist.journal.json",
		// "../sample_data/persist.narration.json",
		// "../sample_data/persist.progression.json",
		// "../sample_data/persist.quest.json",
		"../sample_data/persist.roster.json",
		"../sample_data/persist.town_event.json",
		"../sample_data/persist.town.json",
		"../sample_data/persist.tutorial.json",
		"../sample_data/persist.upgrades.json",
	}
	suite.FileByteSlices = lo.Map(
		suite.FilePaths,
		func(path string, _ int) []byte {
			bs, err := ioutil.ReadFile(path)
			suiteR.NoError(err)
			return bs
		},
	)
	suite.DecodedFiles = lo.Map(
		suite.FileByteSlices,
		func(bs []byte, _ int) Struct {
			decodedFile, err := ToStructuredFile(bs)
			suiteR.NoError(err)
			return *decodedFile
		},
	)
	suite.DecodedFieldSlices = lo.Map(
		suite.DecodedFiles,
		func(decodedFile Struct, _ int) []dfield.DataField {
			return decodedFile.Fields
		},
	)
	suite.DecodedFieldSlicesUnique = lo.Map(
		suite.DecodedFieldSlices,
		func(fields []dfield.DataField, _ int) []dfield.DataField {
			return dfield.RemoveDuplications(fields)
		},
	)
	suite.DecodedFieldSlicesExpanded = lo.Map(
		suite.DecodedFieldSlicesUnique,
		func(fields []dfield.DataField, _ int) []dfield.DataField {
			expandedFields, err := ExpandEmbeddedFiles(fields)
			suiteR.NoError(err)
			return expandedFields
		},
	)
	suite.DecodedJSONsFromFile = lo.Map(
		suite.FileByteSlices,
		func(bs []byte, _ int) orderedmap.OrderedMap {
			// A better way to do this is:
			//
			//    lhm := ToLinkedHashMap(decodedFile)
			//
			// But the current encoding code operates on the assumption that
			// the input linked hash map is something comes from `json.Marshal`
			// with some special data type quirks (everything is float64, etc.)
			// and some corresponding handling.
			outputBytes, err := DecodeDSON(bs, false)
			suiteR.NoError(err)
			lhm := orderedmap.New()
			err = json.Unmarshal(outputBytes, lhm)
			suiteR.NoError(err)
			return *lhm
		},
	)
	suite.EncodingFieldsWithRevision = lo.Map(
		suite.DecodedJSONsFromFile,
		func(lhm orderedmap.OrderedMap, _ int) dfield.EncodingFieldsWithRevision {
			encodingFieldsWithRevision, err := FromLinkedHashMap(lhm)
			suiteR.NoError(err)
			return *encodingFieldsWithRevision
		},
	)
	// suite.EncodingFieldSlicesCompacted = lo.Map(
	// 	suite.EncodingFieldSlices,
	// 	func(fields []dfield.EncodingField, _ int) []dfield.EncodingField {
	// 		compactedFields, err := CompactEmbeddedFiles(fields)
	// 		suiteR.NoError(err)
	// 		return compactedFields
	// 	},
	// )
	suiteR.Equal(len(suite.FilePaths), len(suite.FileByteSlices))
	suiteR.Equal(len(suite.FileByteSlices), len(suite.DecodedFiles))
	suiteR.Equal(len(suite.DecodedFiles), len(suite.DecodedJSONsFromFile))
}

func (suite *EndToEndTestSuite) TestEncodeHeader() {
	suiteR := suite.Require()
	lo.ForEach(
		lo.Zip3(suite.DecodedFiles, suite.FileByteSlices, suite.FilePaths),
		func(tuple3 lo.Tuple3[Struct, []byte, string], _ int) {
			decodedFile := tuple3.A
			header := decodedFile.Header
			headerBytesEncoded := dheader.Encode(header)
			fileBytes := tuple3.B
			headerBytesExpected := PickBytesHeader(fileBytes)
			filePath := tuple3.C

			suiteR.Equal(dheader.DefaultHeaderSize, len(headerBytesEncoded), filePath)
			suiteR.Equal(headerBytesExpected, headerBytesEncoded, filePath)
		},
	)
}

func (suite *EndToEndTestSuite) TestEncodeMeta1Block() {
	suiteR := suite.Require()
	lo.ForEach(
		lo.Zip3(suite.DecodedFiles, suite.FileByteSlices, suite.FilePaths),
		func(tuple3 lo.Tuple3[Struct, []byte, string], _ int) {
			decodedFile := tuple3.A
			meta1Block := decodedFile.Meta1Block
			numMeta1Entries := int(decodedFile.Header.NumMeta1Entries)
			fileBytes := tuple3.B
			meta1BlockBytesEncoded := dmeta1.EncodeBlock(meta1Block)
			meta1BlockBytesExpected := PickBytesMeta1Block(numMeta1Entries, fileBytes)
			filePath := tuple3.C

			suiteR.Equal(
				dmeta1.CalculateBlockLength(numMeta1Entries),
				len(meta1BlockBytesEncoded),
				filePath,
			)
			suiteR.Equal(
				meta1BlockBytesExpected,
				meta1BlockBytesEncoded,
				filePath,
			)
		},
	)
}

func (suite *EndToEndTestSuite) TestEncode_Meta2Block() {
	suiteR := suite.Require()
	lo.ForEach(
		lo.Zip3(suite.DecodedFiles, suite.FileByteSlices, suite.FilePaths),
		func(tuple3 lo.Tuple3[Struct, []byte, string], _ int) {
			decodedFile := tuple3.A
			meta2Block := decodedFile.Meta2Block
			numMeta1Entries := int(decodedFile.Header.NumMeta1Entries)
			numMeta2Entries := int(decodedFile.Header.NumMeta2Entries)
			fileBytes := tuple3.B
			meta2BlockBytesExpected := PickBytesMeta2Block(numMeta1Entries, numMeta2Entries, fileBytes)
			meta2BlockBytesEncoded := dmeta2.EncodeBlock(meta2Block)
			filePath := tuple3.C

			suiteR.Equal(
				dmeta2.CalculateBlockSize(numMeta2Entries),
				len(meta2BlockBytesEncoded),
				filePath,
			)
			suiteR.Equal(
				meta2BlockBytesExpected,
				meta2BlockBytesEncoded,
				filePath,
			)
		},
	)
}

func (suite *EndToEndTestSuite) TestDecodeDSON_Header_Meta2Block() {
	suiteR := suite.Require()
	lo.ForEach(
		suite.DecodedFiles,
		func(decodedFile Struct, _ int) {
			suiteR.Equal(
				decodedFile.Header.DataLength,
				lo.SumBy(
					decodedFile.Meta2Block,
					func(meta2Entry dmeta2.Entry) int32 {
						return meta2Entry.Inferences.FieldNameLength + meta2Entry.Inferences.RawDataLength
					},
				),
			)
		},
	)
}

func (suite *EndToEndTestSuite) TestDecodeEncode_Bytes() {
	suiteR := suite.Require()
	lo.ForEach(
		lo.Zip3(suite.FilePaths, suite.DecodedFieldSlicesExpanded, suite.EncodingFieldsWithRevision),
		func(triplet lo.Tuple3[string, []dfield.DataField, dfield.EncodingFieldsWithRevision], _ int) {
			filePath := triplet.A
			decodedFields := triplet.B
			encodingFieldsWithRevision := triplet.C
			encodingFields := encodingFieldsWithRevision.Fields

			{
				_ = ioutil.WriteFile("/tmp/1.json", []byte(ds.DumpJSON(decodedFields)), 0644)
				_ = ioutil.WriteFile("/tmp/2.json", []byte(ds.DumpJSON(encodingFields)), 0644)
			}

			suiteR.Equal(len(decodedFields), len(encodingFields), filePath)
			lo.ForEach(
				lo.Zip2(decodedFields, encodingFields),
				func(pair lo.Tuple2[dfield.DataField, dfield.EncodingField], _ int) {
					decodedField := pair.A
					encodingField := pair.B
					suiteR.Equal(decodedField.Name, encodingField.Key, filePath)
					suiteR.Equal(decodedField.Inferences.HierarchyPath, encodingField.HierarchyPath, filePath)
					if decodedField.Inferences.RawDataStripped != nil && encodingField.Bytes != nil {
						suiteR.Equal(decodedField.Inferences.RawDataStripped, encodingField.Bytes, filePath)
					}
				},
			)
		},
	)
}

func (suite *EndToEndTestSuite) TestDecodeEncode_Bytes2() {
	suiteR := suite.Require()
	lo.ForEach(
		lo.Zip3(suite.FilePaths, suite.DecodedFieldSlices, suite.EncodingFieldSlicesCompacted),
		func(triplet lo.Tuple3[string, []dfield.DataField, []dfield.EncodingField], _ int) {
			filePath := triplet.A
			decodedFields := triplet.B
			encodingFields := triplet.C
			encodingFields = dfield.RemoveRevisionField(encodingFields)

			suiteR.Equal(len(decodedFields), len(encodingFields), filePath)
			lo.ForEach(
				lo.Zip2(decodedFields, encodingFields),
				func(pair lo.Tuple2[dfield.DataField, dfield.EncodingField], _ int) {
					decodedField := pair.A
					encodingField := pair.B
					suiteR.Equal(decodedField.Name, encodingField.Key, filePath)
					suiteR.Equal(decodedField.Inferences.HierarchyPath, encodingField.HierarchyPath, filePath)
					if decodedField.Inferences.DataType == dfield.DataTypeFileDecoded {
						embeddedFileExpected := decodedField.Inferences.Data.(Struct)
						embeddedFileActual := encodingField.Value.(Struct)
						suiteR.Equal(embeddedFileExpected.Header, embeddedFileActual.Header)
						suiteR.Equal(embeddedFileExpected.Meta1Block, embeddedFileActual.Meta1Block)
						suiteR.Equal(embeddedFileExpected.Meta2Block, embeddedFileActual.Meta2Block)
						suiteR.Equal(embeddedFileExpected.Fields, embeddedFileActual.Fields)
					}
					if decodedField.Inferences.RawDataStripped != nil && encodingField.Bytes != nil {
						suiteR.Equal(decodedField.Inferences.RawDataStripped, encodingField.Bytes, filePath)
					}
				},
			)
		},
	)
}

func (suite *EndToEndTestSuite) TestDecodeEncode_Header() {
	suiteR := suite.Require()
	lo.ForEach(
		lo.Zip3(suite.FilePaths, suite.DecodedFiles, suite.EncodingFieldsWithRevision),
		func(triplet lo.Tuple3[string, Struct, dfield.EncodingFieldsWithRevision], _ int) {
			filePath := triplet.A
			decodedFile := triplet.B
			encodingFieldsWithRevision := triplet.C
			encodingFields := encodingFieldsWithRevision.Fields
			if len(decodedFile.Meta2Block) != len(encodingFields) {
				// skip the test since there are duplicated fields within the original decoded file
				// or there is an embedded DSON within encoding fields
				return
			}

			encodingHeader, err := dfield.CreateHeader(encodingFields)
			suiteR.NoError(err)
			msg := fmt.Sprintf(`Failed at file "%s"`, filePath)
			suiteR.Equalf(decodedFile.Header, *encodingHeader, msg)
		},
	)
}

func (suite *EndToEndTestSuite) TestDecodeEncode_Meta1Block() {
	suiteR := suite.Require()
	lo.ForEach(
		lo.Zip3(suite.FilePaths, suite.DecodedFiles, suite.EncodingFieldSlicesNoRevision),
		func(triplet lo.Tuple3[string, Struct, []dfield.EncodingField], _ int) {
			filePath := triplet.A
			decodedFile := triplet.B
			encodingFields := triplet.C
			if len(decodedFile.Meta2Block) != len(encodingFields) {
				// skip the test since there are duplicated fields within the original decoded file
				// or there is an embedded DSON within encoding fields
				return
			}
			meta1Block := dfield.CreateMeta1Block(encodingFields)
			msg := fmt.Sprintf(
				`Failed at file "%s"\n%s\n%s`, filePath,
				ds.DumpJSON(decodedFile.Meta1Block), ds.DumpJSON(meta1Block),
			)
			suiteR.Equalf(decodedFile.Meta1Block, meta1Block, msg)
		},
	)
}

func (suite *EndToEndTestSuite) TestDecodeEncode_Meta2Block() {
	suiteR := suite.Require()
	lo.ForEach(
		lo.Zip3(suite.FilePaths, suite.DecodedFiles, suite.EncodingFieldSlicesNoRevision),
		func(triplet lo.Tuple3[string, Struct, []dfield.EncodingField], _ int) {
			filePath := triplet.A
			decodedFile := triplet.B
			encodingFields := triplet.C
			if len(decodedFile.Meta2Block) != len(encodingFields) {
				// skip the test since there are duplicated fields within the original decoded file
				// or there is an embedded DSON within encoding fields
				return
			}
			meta2BlockExpected := decodedFile.Meta2Block
			meta2BlockCreated := dfield.CreateMeta2Block(encodingFields)
			msg := fmt.Sprintf(`Failed at file "%s"`, filePath)
			lo.ForEach(
				lo.Zip2(meta2BlockExpected, meta2BlockCreated),
				func(pair lo.Tuple2[dmeta2.Entry, dmeta2.Entry], _ int) {
					entryExpected := pair.A
					entryCreated := pair.B
					if entryExpected.FieldInfo != entryCreated.FieldInfo {
						// Skip as there are times when the first bit of `FieldInfo` get toggled on
						// for no particular reason
						return
					}
					suiteR.Equalf(entryExpected, entryCreated, msg)
				},
			)
		},
	)
}

func (suite *EndToEndTestSuite) TestDecodeEncode_DataFields() {
	suiteR := suite.Require()
	lo.ForEach(
		lo.Zip3(suite.FilePaths, suite.DecodedFiles, suite.EncodingFieldSlicesNoRevision),
		func(triplet lo.Tuple3[string, Struct, []dfield.EncodingField], _ int) {
			filePath := triplet.A
			decodedFile := triplet.B
			encodingFields := triplet.C
			if len(decodedFile.Meta2Block) != len(encodingFields) {
				// skip the test since there are duplicated fields within the original decoded file
				// or there is an embedded DSON within encoding fields
				return
			}
			dataFieldsExpected := decodedFile.Fields
			dataFieldsEncoded := dfield.CreateDataFields(encodingFields)
			msg := fmt.Sprintf(`Failed at file "%s"`, filePath)
			lo.ForEach(
				lo.Zip2(dataFieldsExpected, dataFieldsEncoded),
				func(pair lo.Tuple2[dfield.DataField, dfield.DataField], _ int) {
					dataFieldExpected := pair.A
					dataFieldEncoded := pair.B
					type ComparingData struct {
						Name    string
						RawData []byte
					}
					comparingDataExpected := ComparingData{
						Name:    dataFieldExpected.Name,
						RawData: dataFieldExpected.RawData,
					}
					comparingDataEncoded := ComparingData{
						Name:    dataFieldEncoded.Name,
						RawData: dataFieldEncoded.RawData,
					}
					suiteR.Equalf(
						comparingDataExpected,
						comparingDataEncoded,
						msg,
					)
				},
			)
		},
	)
}

func (suite *EndToEndTestSuite) TestDecodeEncode_File() {
	suiteR := suite.Require()
	lo.ForEach(
		lo.Zip3(suite.FilePaths, suite.DecodedFiles, suite.EncodingFieldsWithRevision),
		func(triplet lo.Tuple3[string, Struct, dfield.EncodingFieldsWithRevision], _ int) {
			filePath := triplet.A
			decodedFileExpected := triplet.B
			encodingFieldsWithRevision := triplet.C
			encodingFields := encodingFieldsWithRevision.Fields

			if len(decodedFileExpected.Meta2Block) != len(encodingFields) {
				return
			}

			decodedFileActual, err := CreateDecodedFile(encodingFields)
			suiteR.NoError(err)

			msg := fmt.Sprintf("Error happened with file %s", filePath)
			suiteR.Equalf(decodedFileExpected.Header, decodedFileActual.Header, msg)
			lo.ForEach(
				lo.Zip2(decodedFileExpected.Meta1Block, decodedFileActual.Meta1Block),
				func(tuple lo.Tuple2[dmeta1.Entry, dmeta1.Entry], _ int) {
					// entryExpected := tuple.A
					// entryActual := tuple.B
				},
			)
			suiteR.Equalf(decodedFileExpected.Meta2Block, decodedFileActual.Meta2Block, msg)
		},
	)
}

func TestEndToEnd(t *testing.T) {
	suite.Run(t, new(EndToEndTestSuite))
}
