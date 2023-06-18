package fimcore

import (
	"errors"

	"github.com/FimGroup/fim/fimapi/basicapi"
	"github.com/FimGroup/fim/fimapi/pluginapi"
)

type Application struct {
	fileManagerMap map[string]pluginapi.FileResourceManager

	sourceConnectorGeneratorMap map[string]pluginapi.SourceConnectorGenerator
	targetConnectorGeneratorMap map[string]pluginapi.TargetConnectorGenerator
}

func (a *Application) Startup() error {
	return nil
}

func (a *Application) Stop() error {
	return nil
}

func (a *Application) AddFileResourceManager(fileManager pluginapi.FileResourceManager) error {
	_, ok := a.fileManagerMap[fileManager.Name()]
	if ok {
		return errors.New("file manager already exists:" + fileManager.Name())
	}
	a.fileManagerMap[fileManager.Name()] = fileManager
	return nil
}

func (a *Application) GetFileResourceManager(name string) pluginapi.FileResourceManager {
	return a.fileManagerMap[name]
}

func (a *Application) AddSourceConnectorGenerator(gen pluginapi.SourceConnectorGenerator) error {
	for _, name := range gen.GeneratorNames() {
		if _, ok := a.sourceConnectorGeneratorMap[name]; ok {
			return errors.New("source connector generator already exists:" + name)
		}
		a.sourceConnectorGeneratorMap[name] = gen
	}
	return nil
}

func (a *Application) AddTargetConnectorGenerator(gen pluginapi.TargetConnectorGenerator) error {
	for _, name := range gen.GeneratorNames() {
		if _, ok := a.targetConnectorGeneratorMap[name]; ok {
			return errors.New("target connector generator already exists:" + name)
		}
		a.targetConnectorGeneratorMap[name] = gen
	}
	return nil
}

func (a *Application) SpawnUseContainer() basicapi.BasicContainer {
	return a.spawnContainer()
}

func (a *Application) spawnContainer() *ContainerInst {
	return newContainer(a)
}

func NewApplication() basicapi.Application {
	return newApplication()
}

func NewPluginApplication() pluginapi.ApplicationSupport {
	return newApplication()
}

func newApplication() *Application {
	return &Application{
		fileManagerMap: map[string]pluginapi.FileResourceManager{},

		sourceConnectorGeneratorMap: map[string]pluginapi.SourceConnectorGenerator{},
		targetConnectorGeneratorMap: map[string]pluginapi.TargetConnectorGenerator{},
	}
}
