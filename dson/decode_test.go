package dson

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"darkest-savior/dson/dfield"
	"darkest-savior/dson/dmeta2"
	"github.com/iancoleman/orderedmap"
	"github.com/samber/lo"
	"github.com/stretchr/testify/suite"
)

type EndToEndTestSuite struct {
	suite.Suite
	FilePaths                  []string
	FileByteSlices             [][]byte
	DecodedFiles               []DecodedFile
	DecodedJSONsFromFile       []orderedmap.OrderedMap
	DecodedFieldSlices         [][]dfield.Field
	DecodedFieldSlicesUnique   [][]dfield.Field
	DecodedFieldSlicesExpanded [][]dfield.Field
	EncodingFieldSlices        [][]dfield.EncodingField
}

func (suite *EndToEndTestSuite) SetupSuite() {
	suiteR := suite.Require()
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
		"../sample_data/persist.progression.json",
		"../sample_data/persist.quest.json",
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
		func(bs []byte, _ int) DecodedFile {
			decodedFile, err := ToStructuredFile(bs)
			suiteR.NoError(err)
			return *decodedFile
		},
	)
	suite.DecodedFieldSlices = lo.Map(
		suite.DecodedFiles,
		func(decodedFile DecodedFile, _ int) []dfield.Field {
			return decodedFile.Fields
		},
	)
	suite.DecodedFieldSlicesUnique = lo.Map(
		suite.DecodedFieldSlices,
		func(fields []dfield.Field, _ int) []dfield.Field {
			return dfield.RemoveDuplications(fields)
		},
	)
	suite.DecodedFieldSlicesExpanded = lo.Map(
		suite.DecodedFieldSlicesUnique,
		func(fields []dfield.Field, _ int) []dfield.Field {
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
	suite.EncodingFieldSlices = lo.Map(
		suite.DecodedJSONsFromFile,
		func(lhm orderedmap.OrderedMap, _ int) []dfield.EncodingField {
			encodingFields, err := FromLinkedHashMap(lhm)
			suiteR.NoError(err)
			return encodingFields
		},
	)
	suiteR.Equal(len(suite.FilePaths), len(suite.FileByteSlices))
	suiteR.Equal(len(suite.FileByteSlices), len(suite.DecodedFiles))
	suiteR.Equal(len(suite.DecodedFiles), len(suite.DecodedJSONsFromFile))
}

func (suite *EndToEndTestSuite) TestDecodeDSON_Header_Meta2Block() {
	suiteR := suite.Require()
	lo.ForEach(
		suite.DecodedFiles,
		func(decodedFile DecodedFile, _ int) {
			suiteR.Equal(
				decodedFile.Header.DataLength,
				lo.SumBy(
					decodedFile.Meta2Block,
					func(meta2Entry dmeta2.Entry) int {
						return meta2Entry.Inferences.FieldNameLength + meta2Entry.Inferences.RawDataLength
					},
				),
			)
		},
	)
}

func (suite *EndToEndTestSuite) TestDecodeEncodeFields() {
	suiteR := suite.Require()
	lo.ForEach(
		lo.Zip2(suite.DecodedFieldSlicesExpanded, suite.EncodingFieldSlices),
		func(pair lo.Tuple2[[]dfield.Field, []dfield.EncodingField], _ int) {
			decodedFields := pair.A
			encodingFields := pair.B
			encodingFields = dfield.RemoveRevisionField(encodingFields)

			suiteR.Equal(len(decodedFields), len(encodingFields))
			lo.ForEach(
				lo.Zip2(decodedFields, encodingFields),
				func(pair lo.Tuple2[dfield.Field, dfield.EncodingField], _ int) {
					decodedField := pair.A
					encodingField := pair.B
					suiteR.Equal(decodedField.Name, encodingField.Key)
					suiteR.Equal(decodedField.Inferences.HierarchyPath, encodingField.HierarchyPath)
					if decodedField.Inferences.RawDataStripped != nil && encodingField.Bytes != nil {
						suiteR.Equal(decodedField.Inferences.RawDataStripped, encodingField.Bytes)
					}
				},
			)
		},
	)
}

func (suite *EndToEndTestSuite) TestDecodeEncodeHeader() {
	suiteR := suite.Require()
	lo.ForEach(
		lo.Zip3(suite.FilePaths, suite.DecodedFiles, suite.EncodingFieldSlices),
		func(triplet lo.Tuple3[string, DecodedFile, []dfield.EncodingField], _ int) {
			filePath := triplet.A
			decodedFile := triplet.B
			encodingFields := triplet.C
			if len(decodedFile.Meta2Block) > len(encodingFields) {
				// skip the test since there are duplicated fields within the original decoded file
				return
			}

			encodingHeader, err := dfield.CreateHeader(encodingFields)
			suiteR.NoError(err)
			msg := fmt.Sprintf(`Failed at file "%s"`, filePath)
			suiteR.Equalf(decodedFile.Header, *encodingHeader, msg)
		},
	)
}

func TestEndToEnd2(t *testing.T) {
	suite.Run(t, new(EndToEndTestSuite))
}
