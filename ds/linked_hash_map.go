package ds

import (
	"bytes"
	"container/list"
	"encoding/json"
)

// LinkedHashMap is a map that remembers insertion-order in deserialization and keys fetching.
type LinkedHashMap[K any, V any] struct {
	hashMap  map[any]any
	ordering *list.List
}

func NewLinkedHashMap[K any, V any]() *LinkedHashMap[K, V] {
	return &LinkedHashMap[K, V]{
		hashMap:  map[any]any{},
		ordering: list.New(),
	}
}

func (r *LinkedHashMap[K, V]) Keys() []K {
	keys := make([]K, 0, r.ordering.Len())
	for runner := r.ordering.Front(); runner != nil; runner = runner.Next() {
		key := runner.Value.(K)
		keys = append(keys, key)
	}
	return keys
}

func (r *LinkedHashMap[K, V]) Put(key K, value V) {
	_, existed := r.hashMap[key]
	// TODO: in case the key existed, check if we need to remove it from ordering
	if !existed {
		r.ordering.PushBack(key)
	}
	r.hashMap[key] = value
}

func (r *LinkedHashMap[K, V]) Get(key K) (V, bool) {
	value, ok := r.hashMap[key].(V)
	return value, ok
}

func (r LinkedHashMap[K, V]) MarshalJSON() ([]byte, error) {
	bs := make([]byte, 0)
	buf := bytes.NewBuffer(bs)

	buf.WriteRune('{')
	for runner := r.ordering.Front(); runner != nil; runner = runner.Next() {
		key := runner.Value.(K)
		value := r.hashMap[key]

		keyBs, err := json.Marshal(key)
		if err != nil {
			return nil, err
		}
		buf.Write(keyBs)

		buf.WriteRune(':')

		valueBs, err := json.Marshal(value)
		if err != nil {
			return nil, err
		}
		buf.Write(valueBs)

		if runner.Next() != nil {
			buf.WriteRune(',')
		}
	}
	buf.WriteRune('}')

	return buf.Bytes(), nil
}
