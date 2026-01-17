package blob

import (
	"github.com/spf13/afero"
)

type (
	StoreOptions struct {
		FS afero.Fs
	}

	Visitor func(*StoreOptions)
)

func (self *StoreOptions) Apply(it Visitor) {
	it(self)
}

func WithRootDir(path string) Visitor {
	return func(so *StoreOptions) {
		real := afero.NewOsFs()
		so.FS = afero.NewBasePathFs(real, path)
	}
}

func WithFS(fs afero.Fs) Visitor {
	return func(so *StoreOptions) {
		so.FS = fs
	}
}
