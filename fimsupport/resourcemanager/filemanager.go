package resourcemanager

import (
	"errors"
	"io"
	"net/http"

	"github.com/FimGroup/fim/fimapi/pluginapi"

	"github.com/spf13/afero"
)

var _ http.FileSystem = new(OsFileResourceManager)

type OsFileResourceManager struct {
	fs   afero.Fs
	name string
}

func (o *OsFileResourceManager) Open(name string) (http.File, error) {
	return afero.NewHttpFs(o.fs).Open(name)
}

func (o *OsFileResourceManager) Startup() error {
	return nil
}

func (o *OsFileResourceManager) Stop() error {
	return nil
}

func (o *OsFileResourceManager) Name() string {
	return o.name
}

func (o *OsFileResourceManager) LoadFile(path string) ([]byte, error) {
	f, err := o.fs.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = f.Close()
	}()
	return io.ReadAll(f)
}

func NewOsFileResourceManager(name, parentPath string) pluginapi.FileResourceManager {
	if name == "" || parentPath == "" {
		panic(errors.New("name or parentPath is empty when creating OsFileResourceManager"))
	}
	return &OsFileResourceManager{
		fs:   afero.NewBasePathFs(afero.NewOsFs(), parentPath),
		name: name,
	}
}
