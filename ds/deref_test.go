package ds

import (
	"testing"

	"github.com/iancoleman/orderedmap"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestDeref(t *testing.T) {
	lhm := orderedmap.New()
	lhm.Set("one", orderedmap.New())
	lhm.Set("two", *orderedmap.New())

	lhm2 := Deref(lhm)

	assert.True(
		t,
		lo.EveryBy(
			lhm2.Keys(),
			func(key string) bool {
				value, _ := lhm2.Get(key)
				_, ok := value.(orderedmap.OrderedMap)
				return ok
			},
		),
	)
}
