package pluginapi

import "github.com/FimGroup/fim/fimapi/basicapi"

type ApplicationSupport interface {
	basicapi.Application

	GetFileResourceManager(name string) FileResourceManager

	AddConfigureManager(manager basicapi.FullConfigureManager) error
	AddFileResourceManager(fileManager FileResourceManager) error
	AddSubConnectorGeneratorDefinitions(tomlData string) error
	AddSourceConnectorGenerator(source SourceConnectorGenerator) error
	AddTargetConnectorGenerator(target TargetConnectorGenerator) error
	AddApplicationListener(listener LifecycleListener)

	Startup() error
	Stop() error
}
