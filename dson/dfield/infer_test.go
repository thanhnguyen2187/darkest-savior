package dfield

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInferHierarchyPath(t *testing.T) {
	createField := func(name string, parentIndex int) Field {
		return Field{
			Name:    name,
			RawData: nil,
			Inferences: Inferences{
				ParentIndex: parentIndex,
			},
		}
	}
	fields := []Field{
		createField("0", -1),
		createField("1", 0),
		createField("2", 0),
		createField("3", 2),
		createField("4", 2),
		createField("5", 2),
		createField("6", 3),
		createField("7", 6),
		createField("8", 6),
	}
	resultMap := map[int][]string{
		0: {"0"},
		1: {"0", "1"},
		3: {"0", "2", "3"},
		7: {"0", "2", "3", "6", "7"},
		8: {"0", "2", "3", "6", "8"},
	}
	for input, expected := range resultMap {
		assert.Equal(t, expected, InferHierarchyPath(input, fields))
	}
}
