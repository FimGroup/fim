package pluginapi

// FileResourceManager defines file resource accessing provider
// Note: more functionality may be provided including:
//   - http.FileSystem
type FileResourceManager interface {
	Name() string
	LoadFile(path string) ([]byte, error)

	Startup() error
	Stop() error
}
