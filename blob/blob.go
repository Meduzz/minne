package blob

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"

	"github.com/Meduzz/helper/fp"
	"github.com/Meduzz/helper/fp/slice"
	"github.com/spf13/afero"
)

type (
	BlobStore interface {
		List(*Query) ([]*BlobEntry, error)
		Read(Object) (io.Reader, error)
		Store(Object, io.Reader) (int64, error)
		Append(Object, io.Reader) (int64, error)
		Remove(Object) error
	}

	blobStore struct {
		options StoreOptions
	}

	BlobEntry struct {
		Name string `json:"name"`
		Dir  bool   `json:"dir"`
		Size int64  `json:"size,omitempty"`
	}

	Query struct {
		Path   string `json:"path,omitempty"`
		Prefix string `json:"prefix,omitempty"`
		Suffix string `json:"suffix,omitempty"`
		Skip   int    `json:"skip,omitempty"`
		Take   int    `json:"take,omitempty"`
	}
)

var (
	_ BlobStore = &blobStore{}
)

func CreateBlobStore(options ...Visitor) (BlobStore, error) {
	config := &StoreOptions{}

	slice.ForEach(options, func(v Visitor) {
		config.Apply(v)
	})

	return &blobStore{
		options: *config,
	}, nil
}

func (b *blobStore) List(query *Query) ([]*BlobEntry, error) {
	ro := afero.NewIOFS(b.options.FS)
	path := "/"

	if query.Path != "" {
		path = query.Path
	}

	entries, err := fs.ReadDir(ro, path)

	if err != nil {
		return nil, err
	}

	if !strings.HasSuffix(path, "/") {
		path = fmt.Sprintf("%s/", path)
	}

	// filter entries
	predicate := query.Predicate()
	entries = slice.Filter(entries, func(de fs.DirEntry) bool {
		return predicate(de.Name())
	})

	// skip
	if query.Skip > 0 {
		entries = slice.Skip(entries, query.Skip)
	}

	// take
	if query.Take > 0 {
		entries = slice.Take(entries, query.Take)
	}

	// map
	return slice.Map(entries, func(it os.DirEntry) *BlobEntry {
		entry := &BlobEntry{}

		if it.IsDir() {
			entry.Name = fmt.Sprintf("%s%s/", path, it.Name())
			entry.Dir = true
		} else {
			entry.Name = fmt.Sprintf("%s%s", path, it.Name())
			entry.Dir = false
			info, _ := it.Info() // iznice

			entry.Size = info.Size()
		}

		return entry
	}), nil
}

func (b *blobStore) Read(obj Object) (io.Reader, error) {
	return b.options.FS.Open(obj.File())
}

func (b *blobStore) Store(obj Object, data io.Reader) (int64, error) {
	file, err := b.options.FS.OpenFile(obj.File(), os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0o644)

	if err != nil {
		return 0, err
	}

	defer file.Close()

	return io.Copy(file, data)
}

func (b *blobStore) Append(obj Object, data io.Reader) (int64, error) {
	file, err := b.options.FS.OpenFile(obj.File(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		return 0, err
	}

	defer file.Close()

	return io.Copy(file, data)
}

func (b *blobStore) Remove(obj Object) error {
	return b.options.FS.Remove(obj.File())
}

func (q *Query) Predicate() fp.Predicate[string] {
	var yes fp.Predicate[string] = func(s string) bool { return true }

	if q.Prefix != "" {
		yes = fp.And(yes, func(s string) bool {
			return strings.HasPrefix(s, q.Prefix)
		})
	}

	if q.Suffix != "" {
		yes = fp.And(yes, func(s string) bool {
			return strings.HasSuffix(s, q.Suffix)
		})
	}

	return yes
}
