package locks

import (
	"io"
	"sync"

	"github.com/Meduzz/minne/blob"
)

type (
	LockSupport interface {
		blob.BlobStore
		Lock(blob.Object)
		Unlock(blob.Object)
	}

	// TODO leave blob.blobStore public for easier inheritence?
	LockStore struct {
		locks map[blob.Object]*sync.Mutex
		store blob.BlobStore
	}
)

var (
	_ LockSupport = &LockStore{}
)

// TODO add a log of unlocked files with timestamp and remove them from map after x minutes.
// TODO will this actually even work in an api? Worst case we're blocking calls until timeouts (you have timeouts right?)
//      Accept a context?

func WithLockSupport(store blob.BlobStore) LockSupport {
	return &LockStore{
		locks: make(map[blob.Object]*sync.Mutex),
		store: store,
	}
}

func (l *LockStore) Lock(object blob.Object) {
	lock, ok := l.locks[object]

	if !ok {
		lock = &sync.Mutex{}
	}

	l.locks[object] = lock

	lock.Lock()
}

func (l *LockStore) Unlock(object blob.Object) {
	lock, ok := l.locks[object]

	if !ok {
		return
	}

	lock.Unlock()
}

func (l *LockStore) List(query *blob.Query) ([]*blob.BlobEntry, error) {
	return l.store.List(query)
}

func (l *LockStore) Read(obj blob.Object) (io.Reader, error) {
	return l.store.Read(obj)
}

func (l *LockStore) Store(obj blob.Object, data io.Reader) (int64, error) {
	return l.store.Store(obj, data)
}

func (l *LockStore) Append(obj blob.Object, data io.Reader) (int64, error) {
	return l.store.Append(obj, data)
}

func (l *LockStore) Remove(obj blob.Object) error {
	return l.store.Remove(obj)
}
