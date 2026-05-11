package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/Meduzz/helper/fp/slice"
	"github.com/Meduzz/minne/blob"
	"github.com/Meduzz/minne/document"
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
	storeItem(l, obj, "one", "Hello one!")
	loadItem(l, obj, "one")
	storeItem(l, obj, "two", "Hello two!")
	loadItem(l, obj, "two")
	storeItem(l, obj, "a", map[string]any{"b": map[string]any{"c": "Woho?"}})
	loadItem(l, obj, "a.b.c")
	storeItem(l, obj, "a.b.c", "Woho!")
	loadItem(l, obj, "a.b.c")
	listItems(l, obj)
	removeItem(l, obj, "two")
	removeItem(l, obj, "a.b.c")
	loadItem(l, obj, "two")
	listItems(l, obj)
	removeItem(l, obj, "one")
	removeItem(l, obj, "a")

	println("\n--- Docs ---")
	doc, _ := blob.NewJsonObject("doc")
	setup(l, doc)
	searchDoc(l, doc, map[string]string{}, 0, 15)
	updateDoc(l, doc, 4, map[string]any{
		"id":    4,
		"name":  "Row 4",
		"value": 0,
		"bonus": "Update",
	})
	readDoc(l, doc, 4)
	removeDoc(l, doc, 9)
	searchDoc(l, doc, map[string]string{}, 0, 15)

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
	data, err := store.Read(object)

	if err != nil {
		println(err.Error())
		os.Exit(1)
	}

	bs, err := io.ReadAll(data)

	if err != nil {
		println(err.Error())
		os.Exit(1)
	}

	items := make(map[string]any)
	err = json.Unmarshal(bs, &items)

	if err != nil {
		println(err.Error())
		os.Exit(1)
	}

	for k, v := range items {
		fmt.Printf("%s=%v\n", k, v)
	}
}

func storeItem(store locks.LockSupport, object blob.Object, key string, value any) {
	err := kv.PatchKey(store, object, key, value)

	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
}

func removeItem(store locks.LockSupport, object blob.Object, key string) {
	err := kv.RemoveKey(store, object, key)

	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
}

func loadItem(store locks.LockSupport, object blob.Object, key string) {
	item, err := kv.LoadKey(store, object, key)

	if err != nil {
		println(err.Error())
		os.Exit(1)
	}

	fmt.Printf("%v\n", item)
}

func setup(store locks.LockSupport, object blob.Object) {
	if object == "" {
		println("no object")
	}

	i := 0
	for i < 10 {
		data := map[string]any{
			"id":    i,
			"name":  fmt.Sprintf("Row %d", i),
			"value": i % 2,
		}

		err := document.Create(store, object, data)

		if err != nil {
			println(err.Error())
			os.Exit(1)
		}

		i++
	}
}

func searchDoc(store locks.LockSupport, object blob.Object, filter map[string]string, skip, take int) {
	result, err := document.List(store, object, filter, skip, take)

	if err != nil {
		println(err.Error())
		os.Exit(1)
	}

	slice.ForEach(result, func(it map[string]any) {
		fmt.Printf("%v\n", it)
	})
}

func updateDoc(store locks.LockSupport, object blob.Object, id int, val map[string]any) {
	err := document.Update(store, object, "id", id, val)

	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
}

func readDoc(store locks.LockSupport, object blob.Object, id int) {
	item, err := document.Read(store, object, "id", id)

	if err != nil {
		println(err.Error())
		os.Exit(1)
	}

	fmt.Printf("%v\n", item)
}

func removeDoc(store locks.LockSupport, object blob.Object, id int) {
	err := document.Delete(store, object, "id", id)

	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
}
