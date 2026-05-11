package document

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/Meduzz/helper/fp/slice"
	"github.com/Meduzz/minne/blob"
	"github.com/Meduzz/minne/locks"
	"github.com/Meduzz/minne/store"
)

type (
	mutation func([]map[string]any) []map[string]any
	callback func([]map[string]any)
)

func Create(blobStore locks.LockSupport, obj blob.Object, doc map[string]any) error {
	return store.MutateObject(blobStore, obj, func(r io.Reader) (io.Reader, error) {
		return mutate(r, func(m []map[string]any) []map[string]any {
			return append(m, doc)
		})
	}, "[]")
}

func Update(blobStore locks.LockSupport, obj blob.Object, key string, value any, doc map[string]any) error {
	return store.MutateObject(blobStore, obj, func(r io.Reader) (io.Reader, error) {
		return mutate(r, func(m []map[string]any) []map[string]any {
			return slice.Map(m, func(it map[string]any) map[string]any {
				if strings.Contains(key, ".") {
					keys := strings.Split(key, ".")
					edge, k := lens(it, keys)

					if match(edge[k], value) {
						return doc
					} else {
						return it
					}
				} else {
					if match(it[key], value) {
						return doc
					} else {
						return it
					}
				}
			})
		})
	}, "[]")
}

func Read(blobStore locks.LockSupport, obj blob.Object, key string, value any) ([]map[string]any, error) {
	var matches []map[string]any

	err := store.ReadObject(blobStore, obj, func(r io.Reader) error {
		return read(r, func(m []map[string]any) {
			matches = slice.Filter(m, func(it map[string]any) bool {
				if strings.Contains(key, ".") {
					keys := strings.Split(key, ".")
					edge, k := lens(it, keys)

					return match(edge[k], value)
				} else {
					return match(it[key], value)
				}
			})
		})
	}, "[]")

	return matches, err
}

func Delete(blobStore locks.LockSupport, obj blob.Object, key string, value any) error {
	return store.MutateObject(blobStore, obj, func(r io.Reader) (io.Reader, error) {
		return mutate(r, func(m []map[string]any) []map[string]any {
			return slice.Filter(m, func(it map[string]any) bool {
				if strings.Contains(key, ".") {
					keys := strings.Split(key, ".")
					edge, k := lens(it, keys)

					return match(edge[k], value)
				} else {
					return match(it[key], value)
				}
			})

		})
	}, "[]")
}

func List(blobStore locks.LockSupport, obj blob.Object, filter map[string]string, skip, take int) ([]map[string]any, error) {
	var matches []map[string]any

	err := store.ReadObject(blobStore, obj, func(r io.Reader) error {
		return read(r, func(m []map[string]any) {
			if len(filter) == 0 {
				matches = m
			} else {
				matches = slice.Filter(m, func(it map[string]any) bool {
					var result []bool

					for key, value := range filter {
						if strings.Contains(key, ".") {
							keys := strings.Split(key, ".")
							edge, k := lens(it, keys)

							real, ok := edge[k]

							if !ok {
								result = append(result, false)
								continue
							}

							result = append(result, match(real, value))
						} else {
							real, ok := it[key]

							if !ok {
								result = append(result, false)
								continue
							}

							result = append(result, match(real, value))
						}
					}

					return slice.Fold(result, true, func(it, agg bool) bool {
						if !agg {
							return agg
						} else {
							return it
						}
					})
				})
			}
		})
	}, "[]")

	return slice.Take(slice.Skip(matches, skip), take), err
}

func mutate(in io.Reader, cb mutation) (io.Reader, error) {
	bs, err := io.ReadAll(in)

	if err != nil {
		return nil, err
	}

	data := make([]map[string]any, 0)

	err = json.Unmarshal(bs, &data)

	if err != nil {
		return nil, err
	}

	update := cb(data)

	bs, err = json.Marshal(update)

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

	data := make([]map[string]any, 0)

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

func match(first, second any) bool {
	return fmt.Sprintf("%v", first) == fmt.Sprintf("%v", second)
}
