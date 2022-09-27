package dfield

import (
	"encoding/json"
	"testing"

	"github.com/iancoleman/orderedmap"
	"github.com/stretchr/testify/require"
	"github.com/thanhnguyen2187/darkest-savior/ds"
)

func TestCalculateNumChildren(t *testing.T) {
	lhm := orderedmap.New()
	jsonStr := `
{
  "0": "zero",
  "1": "one",
  "2": {},
  "3": {
    "4": "four",
    "5": {
      "6": "six",
      "7": "seven"
    },
    "8": "eight"
  },
  "9": "nine"
}
`
	err := json.Unmarshal([]byte(jsonStr), lhm)
	require.NoError(t, err)
	ds.Deref(lhm)

	{
		numsDirectChildren := CalculateNumDirectChildren(*lhm)
		expected := []int{0, 0, 0, 3, 0, 2, 0, 0, 0, 0}
		require.Equal(t, expected, numsDirectChildren)
	}
	{
		numAllChildren := CalculateNumAllChildren(*lhm)
		expected := []int{0, 0, 0, 5, 0, 2, 0, 0, 0, 0}
		require.Equal(t, expected, numAllChildren)
	}
}

func TestCalculateParentIndexes(t *testing.T) {
	numsAllChildren := []int{0, 0, 6, 0, 0, 0, 2, 0, 0, 1, 0}
	parentIndexes := CalculateParentIndexes(numsAllChildren)
	expected := []int{-1, -1, -1, 2, 2, 2, 2, 6, 6, -1, 9}
	require.Equal(t, expected, parentIndexes)
}
