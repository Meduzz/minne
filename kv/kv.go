package kv

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/Meduzz/minne/blob"
	"github.com/Meduzz/minne/locks"
	"github.com/Meduzz/minne/store"
)

type (
	callback = func(map[string]any)
)

func Store(blobStore locks.LockSupport, obj blob.Object, key string, value any) error {
	return store.MutateObject(blobStore, obj, func(in io.Reader) (io.Reader, error) {
		return mutate(in, func(data map[string]any) {
			data[key] = value
		})
	})
}

func Remove(blobStore locks.LockSupport, obj blob.Object, key string) error {
	return store.MutateObject(blobStore, obj, func(in io.Reader) (io.Reader, error) {
		return mutate(in, func(data map[string]any) {
			delete(data, key)
		})
	})
}

func Load(blobStore locks.LockSupport, obj blob.Object, key string) (any, error) {
	var item any

	err := store.ReadObject(blobStore, obj, func(in io.Reader) error {
		return read(in, func(data map[string]any) {
			item = data[key]
		})
	})

	return item, err
}

func List(blobStore locks.LockSupport, obj blob.Object) (map[string]any, error) {
	var item map[string]any

	err := store.ReadObject(blobStore, obj, func(in io.Reader) error {
		return read(in, func(data map[string]any) {
			item = data
		})
	})

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
