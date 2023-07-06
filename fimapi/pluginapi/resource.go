package pluginapi

type FileResourceManager interface {
	Name() string
	LoadFile(path string) ([]byte, error)

	Startup() error
	Stop() error
}
