package kv

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"

	"github.com/Meduzz/helper/fp/slice"
	"github.com/Meduzz/minne/blob"
	"github.com/Meduzz/minne/locks"
	"github.com/Meduzz/minne/store"
)

type (
	callback = func(map[string]any)
)

func PatchKey(blobStore locks.LockSupport, obj blob.Object, key string, value any) error {
	return store.MutateObject(blobStore, obj, func(in io.Reader) (io.Reader, error) {
		return mutate(in, func(data map[string]any) {
			if strings.Contains(key, ".") {
				keys := strings.Split(key, ".")
				edge, k := lens(data, keys)
				edge[k] = value
			} else {
				data[key] = value
			}
		})
	}, "")
}

func RemoveKey(blobStore locks.LockSupport, obj blob.Object, key string) error {
	return store.MutateObject(blobStore, obj, func(in io.Reader) (io.Reader, error) {
		return mutate(in, func(data map[string]any) {
			if strings.Contains(key, ".") {
				keys := strings.Split(key, ".")
				edge, k := lens(data, keys)
				delete(edge, k)
			} else {
				delete(data, key)
			}
		})
	}, "")
}

func LoadKey(blobStore locks.LockSupport, obj blob.Object, key string) (any, error) {
	var item any

	err := store.ReadObject(blobStore, obj, func(in io.Reader) error {
		return read(in, func(data map[string]any) {
			if strings.Contains(key, ".") {
				keys := strings.Split(key, ".")
				edge, k := lens(data, keys)
				item = edge[k]
			} else {
				item = data[key]
			}
		})
	}, "")

	return item, err
}

func mutate(in io.Reader, cb callback) (io.Reader, error) {
	bs, err := io.ReadAll(in)

	if err != nil {
		return nil, err
	}

	var data map[string]any

	err = json.Unmarshal(bs, &data)

	if err != nil {
		return nil, err
	}

	cb(data)

	bs, err = json.Marshal(data)

	if err != nil {
		return nil, err
	}

	return bytes.NewReader(bs), nil
}

func read(in io.Reader, cb callback) error {
	bs, err := io.ReadAll(in)

	if err != nil {
		return err
	}

	var data map[string]any

	err = json.Unmarshal(bs, &data)

	if err != nil {
		return err
	}

	cb(data)

	return nil
}

func lens(data map[string]any, key []string) (map[string]any, string) {
	// fetch first key
	head := slice.Head(key)
	// fetch data in map at first key
	item, ok := data[head]

	// it's not there
	if !ok {
		// create it
		item = make(map[string]any)
		data[head] = item
	}

	// cast item to map
	obj, ok := item.(map[string]any)

	if !ok {
		// overwrite
		data[head] = make(map[string]any)
	}

	// fetch tail of slice
	tail := slice.Tail(key)

	// if one item, left return it
	if len(tail) == 1 {
		return obj, tail[0]
	}

	// otherweise recurse
	return lens(obj, tail)
}
