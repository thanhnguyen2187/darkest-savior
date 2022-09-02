package dhash

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashString(t *testing.T) {
	expectedValues := map[string]int32{
		"":              0,
		"crusader":      1181166609,
		"plague_doctor": -586237712,
	}
	for s, i := range expectedValues {
		assert.Equal(t, HashString(s), i)
	}
}
