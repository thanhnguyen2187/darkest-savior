package dson

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"darkest-savior/ds"
	"darkest-savior/dson/dfield"
	"github.com/iancoleman/orderedmap"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
)

func TestEndToEnd(t *testing.T) {
	fileNames := []string{
		// "../sample_data/novelty_tracker.json",
		"../sample_data/persist.campaign_log.json",
		// "../sample_data/persist.campaign_mash.json",
		// "../sample_data/persist.curio_tracker.json",
		// "../sample_data/persist.estate.json",
		// "../sample_data/persist.game.json",
		// "../sample_data/persist.game_knowledge.json",
		// "../sample_data/persist.journal.json",
		// "../sample_data/persist.narration.json",
		// "../sample_data/persist.progression.json",
		// "../sample_data/persist.quest.json",
		// "../sample_data/persist.roster.json",
		// "../sample_data/persist.town_event.json",
		// "../sample_data/persist.town.json",
		// "../sample_data/persist.tutorial.json",
		// "../sample_data/persist.upgrades.json",
	}
	for _, fileName := range fileNames {

		inputBytes, err := ioutil.ReadFile(fileName)
		require.NoError(t, err)

		decodedFile := DecodedFile{}
		{
			outputBytes, err := DecodeDSON(inputBytes, true)
			require.NoError(t, err)
			err = json.Unmarshal(outputBytes, &decodedFile)
			require.NoError(t, err)
		}
		lhm := orderedmap.New()
		{
			outputBytes, err := DecodeDSON(inputBytes, false)
			require.NoError(t, err)
			err = json.Unmarshal(outputBytes, &lhm)
			require.NoError(t, err)
		}

		// `uniqueFields` is needed since decodedFile.Fields has duplicating names sometimes
		// and the underlying data structure (linked hash map) just kind of... swallow that up.
		// One quite "cool" theory is that the underlying data structure of Darkest Dungeon's developers
		// is something similar to a linked hash map that has the hash map itself good,
		// but the self implemented linked list does not work as expected.
		//
		// Also see: https://github.com/robojumper/DarkestDungeonSaveEditor/issues/11
		uniqueFields := dfield.RemoveDuplications(decodedFile.Fields)
		// uniqueFields := decodedFile.Fields
		dataFields := uniqueFields
		dataFields = lo.FlatMap(
			dataFields,
			func(field dfield.Field, _ int) []dfield.Field {
				if field.Inferences.DataType == dfield.DataTypeFileDecoded {
					decodedFileBytes, err := json.Marshal(field.Inferences.Data)
					require.NoError(t, err)

					decodedFile := DecodedFile{}
					err = json.Unmarshal(decodedFileBytes, &decodedFile)
					require.NoError(t, err)

					return append(
						[]dfield.Field{field},
						decodedFile.Fields...,
					)
				}
				return []dfield.Field{field}
			},
		)
		encodingFields, err := FromLinkedHashMap(*lhm)
		require.NoError(t, err)
		encodingFields = lo.Filter(
			encodingFields,
			func(field dfield.EncodingField, _ int) bool {
				return field.Key != "__revision_dont_touch"
			},
		)

		{
			data1 := lo.Map(
				dataFields,
				func(field dfield.Field, _ int) []string {
					return field.Inferences.HierarchyPath
				},
			)
			data2 := lo.Map(
				encodingFields,
				func(field dfield.EncodingField, _ int) []string {
					return field.HierarchyPath
				},
			)
			_ = ioutil.WriteFile("/tmp/1.txt", []byte(ds.DumpJSON(data1)), 0644)
			_ = ioutil.WriteFile("/tmp/2.txt", []byte(ds.DumpJSON(data2)), 0644)
		}

		require.Equal(t, len(dataFields), len(encodingFields))
		for i, pair := range lo.Zip2(dataFields, encodingFields) {
			require.Equalf(t, pair.A.Name, pair.B.Key, "%d %s", i, ds.DumpJSON(pair))
			require.Equalf(t, pair.A.Inferences.HierarchyPath, pair.B.HierarchyPath, "%d %s", i, ds.DumpJSON(pair))
			if pair.A.Inferences.RawDataStripped != nil && pair.B.Bytes != nil {
				require.Equalf(t, pair.A.Inferences.RawDataStripped, pair.B.Bytes, "%d %s", i, ds.DumpJSON(pair))
			}
		}

	}
}
