package ds

import (
	"github.com/iancoleman/orderedmap"
)

func Deref(lhm *orderedmap.OrderedMap) orderedmap.OrderedMap {
	for _, key := range lhm.Keys() {
		value, _ := lhm.Get(key)
		valueOM, ok := value.(*orderedmap.OrderedMap)
		if ok {
			for _, key := range lhm.Keys() {
				lhm.Set(key, Deref(valueOM))
			}
		}
	}
	return *lhm
}
