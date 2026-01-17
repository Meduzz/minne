package main

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/Meduzz/helper/fp/slice"
	"github.com/Meduzz/minne/blob"
	"github.com/Meduzz/minne/kv"
	"github.com/Meduzz/minne/locks"
	"github.com/spf13/afero"
)

func main() {
	b, _ := blob.CreateBlobStore(blob.WithFS(afero.NewMemMapFs()))

	println("-- Files --")
	listFiles(b, "/")
	storeFile(b, "/test.txt", "Hello world")
	readFile(b, "/test.txt")
	appendFile(b, "/test.txt", "\n\nAnd a happy new year!")
	readFile(b, "/test.txt")

	println("\n-- List ---")
	obj, _ := blob.NewJsonObject("test")
	l := locks.WithLockSupport(b)
	listItems(l, obj)
	storeItem(l, obj, "one", "Hello one!")
	loadItem(l, obj, "one")
	storeItem(l, obj, "two", "Hello two!")
	loadItem(l, obj, "two")
	listItems(l, obj)
	removeItem(l, obj, "two")
	loadItem(l, obj, "two")
	listItems(l, obj)
	removeItem(l, obj, "one")

	println("\n-- Back in files ---")
	listFiles(b, "/")
	readFile(b, "/test.json")
	removeFile(b, "/test.json")
	listFiles(b, "/")
}

func appendFile(store blob.BlobStore, fileName string, data string) {
	add := bytes.NewBufferString(data)

	_, err := store.Append(blob.Object(fileName), add)

	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
}

func storeFile(store blob.BlobStore, fileName string, data string) {
	buf := bytes.NewBufferString(data)

	_, err := store.Store(blob.Object(fileName), buf)

	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
}

func readFile(store blob.BlobStore, fileName string) {
	data, err := store.Read(blob.Object(fileName))

	if err != nil {
		println(err.Error())
		os.Exit(1)
	}

	bs, err := io.ReadAll(data)

	if err != nil {
		println(err.Error())
		os.Exit(1)
	}

	println(string(bs))
}

func listFiles(store blob.BlobStore, path string) {
	entries, err := store.List(&blob.Query{Path: path})

	if err != nil {
		println(err.Error())
		os.Exit(1)
	}

	slice.ForEach(entries, func(it *blob.BlobEntry) {
		fmt.Printf("%s %d\n", it.Name, it.Size)
	})
}

func removeFile(store blob.BlobStore, fileName string) {
	err := store.Remove(blob.Object(fileName))

	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
}

func listItems(store locks.LockSupport, object blob.Object) {
	items, _ := kv.List(store, object)

	for k, v := range items {
		fmt.Printf("%s=%v\n", k, v)
	}
}

func storeItem(store locks.LockSupport, object blob.Object, key string, value any) {
	err := kv.Store(store, object, key, value)

	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
}

func removeItem(store locks.LockSupport, object blob.Object, key string) {
	err := kv.Remove(store, object, key)

	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
}

func loadItem(store locks.LockSupport, object blob.Object, key string) {
	item, err := kv.Load(store, object, key)

	if err != nil {
		println(err.Error())
		os.Exit(1)
	}

	fmt.Printf("%v\n", item)
}
