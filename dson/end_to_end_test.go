package dson

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"darkest-savior/dson/dfield"
	"darkest-savior/dson/dmeta1"
	"darkest-savior/dson/dmeta2"
	"github.com/iancoleman/orderedmap"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type EndToEndTestSuite2 struct {
	FilePaths            []string
	FileByteSlices       [][]byte
	DecodedStructs       []Struct
	DecodedJSONsFromFile []orderedmap.OrderedMap
	EncodingStructs      []Struct
	R                    *require.Assertions
	suite.Suite
}

func (suite *EndToEndTestSuite2) SetupSuite() {
	suite.R = suite.Require()
	suite.FilePaths = []string{
		"../sample_data/novelty_tracker.json",
		"../sample_data/persist.campaign_log.json",
		"../sample_data/persist.campaign_mash.json",
		"../sample_data/persist.curio_tracker.json",
		"../sample_data/persist.estate.json",
		"../sample_data/persist.game.json",
		"../sample_data/persist.game_knowledge.json",
		"../sample_data/persist.journal.json",
		"../sample_data/persist.narration.json",
		// "../sample_data/persist.progression.json", // has duplicated field
		"../sample_data/persist.quest.json",
		"../sample_data/persist.roster.json", // has embedded DSON file
		"../sample_data/persist.town_event.json",
		"../sample_data/persist.town.json",
		"../sample_data/persist.tutorial.json",
		"../sample_data/persist.upgrades.json",
	}
	suite.FileByteSlices = lo.Map(
		suite.FilePaths,
		func(path string, _ int) []byte {
			bs, err := ioutil.ReadFile(path)
			suite.R.NoError(err)
			return bs
		},
	)
	suite.DecodedStructs = lo.Map(
		suite.FileByteSlices,
		func(bs []byte, _ int) Struct {
			decodedFile, err := ToStructuredFile(bs)
			suite.R.NoError(err)
			return *decodedFile
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
			suite.R.NoError(err)
			lhm := orderedmap.New()
			err = json.Unmarshal(outputBytes, lhm)
			suite.R.NoError(err)
			return *lhm
		},
	)
	suite.EncodingStructs = lo.Map(
		suite.DecodedJSONsFromFile,
		func(lhm orderedmap.OrderedMap, _ int) Struct {
			encodingStruct, err := FromLinkedHashMapV2(lhm)
			suite.R.NoError(err)
			return *encodingStruct
		},
	)
}

func (suite *EndToEndTestSuite2) TestMeta2Block(filePath string, expected []dmeta2.Entry, actual []dmeta2.Entry) {
	suite.R.Equal(len(expected), len(actual))
	lo.ForEach(
		lo.Zip2(expected, actual),
		func(tuple lo.Tuple2[dmeta2.Entry, dmeta2.Entry], _ int) {
			expected := tuple.A
			actual := tuple.B
			suite.R.Equalf(expected.NameHash, actual.NameHash, filePath)
			suite.R.Equalf(expected.Offset, actual.Offset, filePath)
			transform := func(fieldInfo int32) uint32 {
				bits := uint32(0b01111111111111111111111111111111)
				return uint32(fieldInfo) & bits
			}
			suite.R.Equalf(
				transform(expected.FieldInfo),
				transform(actual.FieldInfo),
				filePath,
			)
		},
	)
}

func (suite *EndToEndTestSuite2) TestEmbeddedDataFields(filePath string, expected []dfield.DataField, actual []dfield.DataField) {
	suite.R.Equal(len(expected), len(actual))
	lo.ForEach(
		lo.Zip2(expected, actual),
		func(tuple lo.Tuple2[dfield.DataField, dfield.DataField], _ int) {
			expected := tuple.A
			actual := tuple.B
			suite.R.Equalf(expected.Name, actual.Name, filePath)
			// Do not check RawData directly, it contains padded zeroes with the actual data
			// and sometimes, the padded zeroes are not actual padded zeroes.
			//
			// Also see: https://github.com/robojumper/DarkestDungeonSaveEditor/issues/51
			suite.R.Equalf(len(expected.RawData), len(actual.RawData), filePath)
			suite.R.Equalf(
				expected.Inferences.RawDataStripped,
				actual.Inferences.RawDataStripped,
				filePath,
			)
		},
	)
}

func (suite *EndToEndTestSuite2) TestEncodeDecode_DataFields() {
	lo.ForEach(
		lo.Zip3(suite.FilePaths, suite.DecodedStructs, suite.EncodingStructs),
		func(tuple lo.Tuple3[string, Struct, Struct], _ int) {
			filePath := tuple.A
			decodedStruct := tuple.B
			encodingStruct := tuple.C
			suite.R.Equalf(len(decodedStruct.Fields), len(encodingStruct.Fields), filePath)
			lo.ForEach(
				lo.Zip2(decodedStruct.Fields, encodingStruct.Fields),
				func(tuple lo.Tuple2[dfield.DataField, dfield.DataField], _ int) {
					fieldExpected := tuple.A
					fieldActual := tuple.B
					embeddedStructExpected, ok1 := fieldExpected.Inferences.Data.(Struct)
					embeddedStructActual, ok2 := fieldActual.Inferences.Data.(Struct)
					if ok1 && ok2 {
						suite.R.Equalf(embeddedStructExpected.Header, embeddedStructActual.Header, filePath)
						suite.R.Equalf(embeddedStructExpected.Meta1Block, embeddedStructActual.Meta1Block, filePath)
						suite.TestMeta2Block(filePath, embeddedStructExpected.Meta2Block, embeddedStructActual.Meta2Block)
						suite.TestEmbeddedDataFields(filePath, embeddedStructExpected.Fields, embeddedStructActual.Fields)
					} else if !ok1 && !ok2 {
						suite.R.Equalf(fieldExpected.Name, fieldActual.Name, filePath)
						suite.R.Equalf(fieldExpected.Inferences.RawDataStripped, fieldActual.Inferences.RawDataStripped, filePath)
						suite.R.Equalf(fieldExpected.Inferences.HierarchyPath, fieldActual.Inferences.HierarchyPath, filePath)
						suite.R.Equalf(fieldExpected.RawData, fieldActual.RawData, filePath)
					} else {
						suite.Fail("Unreachable code")
					}
				},
			)
		},
	)
}

func (suite *EndToEndTestSuite2) TestEncodeDecode_Meta2Block() {
	lo.ForEach(
		lo.Zip3(suite.FilePaths, suite.DecodedStructs, suite.EncodingStructs),
		func(tuple lo.Tuple3[string, Struct, Struct], _ int) {
			filePath := tuple.A
			decodedStruct := tuple.B
			encodingStruct := tuple.C
			suite.TestMeta2Block(
				filePath,
				decodedStruct.Meta2Block,
				encodingStruct.Meta2Block,
			)
		},
	)
}

func (suite *EndToEndTestSuite2) TestEncodeDecode_Meta1Block() {
	lo.ForEach(
		lo.Zip3(suite.FilePaths, suite.DecodedStructs, suite.EncodingStructs),
		func(tuple lo.Tuple3[string, Struct, Struct], _ int) {
			filePath := tuple.A
			decodedStruct := tuple.B
			encodingStruct := tuple.C
			suite.R.Equalf(len(decodedStruct.Meta1Block), len(encodingStruct.Meta1Block), filePath)
			lo.ForEach(
				lo.Zip2(decodedStruct.Meta1Block, encodingStruct.Meta1Block),
				func(tuple lo.Tuple2[dmeta1.Entry, dmeta1.Entry], _ int) {
					meta1BlockExpected := tuple.A
					meta1BlockActual := tuple.B
					suite.R.Equalf(meta1BlockExpected, meta1BlockActual, filePath)
				},
			)
		},
	)
}

func (suite *EndToEndTestSuite2) TestEncodeDecode_Header() {
	lo.ForEach(
		lo.Zip3(suite.FilePaths, suite.DecodedStructs, suite.EncodingStructs),
		func(tuple lo.Tuple3[string, Struct, Struct], _ int) {
			filePath := tuple.A
			decodedStruct := tuple.B
			encodingStruct := tuple.C
			suite.R.Equalf(decodedStruct.Header, encodingStruct.Header, filePath)
		},
	)
}

func (suite *EndToEndTestSuite2) TestEncodeDecode_Struct() {
	lo.ForEach(
		lo.Zip3(suite.FilePaths, suite.FileByteSlices, suite.DecodedStructs),
		func(tuple lo.Tuple3[string, []byte, Struct], _ int) {
			filePath := tuple.A
			fileBytes := tuple.B
			decodedStruct := tuple.C
			decodedStructBytes := EncodeStruct(decodedStruct)

			suite.R.Equalf(len(fileBytes), len(decodedStructBytes), filePath)
			suite.R.Equalf(fileBytes, decodedStructBytes, filePath)
		},
	)
}

func TestEndToEnd2(t *testing.T) {
	suite.Run(t, new(EndToEndTestSuite2))
}
