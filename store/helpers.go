package store

import (
	"bytes"
	"errors"
	"os"

	"github.com/Meduzz/minne/blob"
	"github.com/Meduzz/minne/locks"
)

func MutateObject(store locks.LockSupport, obj blob.Object, cb MutateObjectCallback) error {
	store.Lock(obj)
	defer store.Unlock(obj)

	stream, err := store.Read(obj)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			stream = bytes.NewReader([]byte("{}"))
		} else {
			return err
		}
	}

	updated, err := cb(stream)

	// TODO do something with bytes written....
	_, err = store.Store(obj, updated)

	if err != nil {
		return err
	}

	return nil
}

func ReadObject(store locks.LockSupport, obj blob.Object, cb ReadObjectCallback) error {
	stream, err := store.Read(obj)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			stream = bytes.NewReader([]byte("{}"))
		} else {
			return err
		}
	}

	return cb(stream)
}
