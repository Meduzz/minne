package store

import "io"

type (
	MutateObjectCallback func(io.Reader) (io.Reader, error)
	ReadObjectCallback   func(io.Reader) error
)
