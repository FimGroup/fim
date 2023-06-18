package pluginapi

import "github.com/FimGroup/fim/fimapi/basicapi"

type ApplicationSupport interface {
	basicapi.Application

	GetFileResourceManager(name string) FileResourceManager
	AddFileResourceManager(fileManager FileResourceManager) error
}
