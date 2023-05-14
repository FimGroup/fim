package fimcore

import (
	"io"

	"github.com/spf13/afero"
)

func LoadFlowModelFile(file string) ([]byte, error) {
	fs := afero.NewOsFs()
	f, err := fs.Open(file)
	if err != nil {
		return nil, err
	}
	defer func(f afero.File) {
		_ = f.Close()
	}(f)

	return io.ReadAll(f)
}
