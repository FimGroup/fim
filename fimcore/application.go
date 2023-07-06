package fimcore

import (
	"errors"
	"strings"

	"github.com/FimGroup/fim/fimapi/basicapi"
	"github.com/FimGroup/fim/fimapi/pluginapi"
)

type Application struct {
	fileManagerMap       map[string]pluginapi.FileResourceManager
	configureManagerList []basicapi.FullConfigureManager
	configureManager     *NestedConfigureManager
	lifecycleListeners   []pluginapi.LifecycleListener

	sourceConnectorGeneratorMap map[string]pluginapi.SourceConnectorGenerator
	targetConnectorGeneratorMap map[string]pluginapi.TargetConnectorGenerator

	sourceConnectorGeneratorDefinitions    []map[string]map[string]string
	targetConnectorGeneratorDefinitions    []map[string]map[string]string
	preInitializedSourceConnectorGenerator map[string]pluginapi.SourceConnectorGenerator
	preInitializedTargetConnectorGenerator map[string]pluginapi.TargetConnectorGenerator

	stopFn func() error
}

func (a *Application) AddApplicationListener(listener pluginapi.LifecycleListener) {
	a.lifecycleListeners = append(a.lifecycleListeners, listener)
}

func (a *Application) AddConfigureManager(manager basicapi.FullConfigureManager) error {
	if manager != nil {
		a.configureManagerList = append(a.configureManagerList, manager)
		a.configureManager.addSubConfigureManager(manager)
	}
	return nil
}

func (a *Application) Startup() error {
	// init configure manager
	for _, v := range a.configureManagerList {
		if err := v.Startup(); err != nil {
			return err
		}
	}
	// init file manager
	for _, v := range a.fileManagerMap {
		if err := v.Startup(); err != nil {
			return err
		}
	}
	// init source and target connector
	for _, v := range a.targetConnectorGeneratorMap {
		if err := v.Startup(); err != nil {
			return err
		}
	}
	for _, v := range a.sourceConnectorGeneratorMap {
		if err := v.Startup(); err != nil {
			return err
		}
	}
	// init source and target connector generator
	for _, v := range a.sourceConnectorGeneratorDefinitions {
		if err := a.setupAndStoreSubSourceConnectorGenerator(v); err != nil {
			return err
		}
	}
	for _, v := range a.targetConnectorGeneratorDefinitions {
		if err := a.setupAndStoreSubTargetConnectorGenerator(v); err != nil {
			return err
		}
	}
	for _, v := range a.preInitializedTargetConnectorGenerator {
		if err := v.Startup(); err != nil {
			return err
		}
	}
	for _, v := range a.preInitializedSourceConnectorGenerator {
		if err := v.Startup(); err != nil {
			return err
		}
	}
	// trigger lifecycle listener at end
	for _, v := range a.lifecycleListeners {
		if err := v.OnStart(); err != nil {
			return err
		}
	}

	a.stopFn = func() error {
		// reverse order to Startup function
		// trigger lifecycle listener at start
		for _, v := range a.lifecycleListeners {
			if err := v.OnStop(); err != nil {
				return err
			}
		}
		// stop source and target connector generator
		for _, v := range a.preInitializedSourceConnectorGenerator {
			if err := v.Stop(); err != nil {
				return err
			}
		}
		for _, v := range a.preInitializedTargetConnectorGenerator {
			if err := v.Stop(); err != nil {
				return err
			}
		}
		// stop source and target connector
		for _, v := range a.sourceConnectorGeneratorMap {
			if err := v.Stop(); err != nil {
				return err
			}
		}
		for _, v := range a.targetConnectorGeneratorMap {
			if err := v.Stop(); err != nil {
				return err
			}
		}
		// stop file manager
		for _, v := range a.fileManagerMap {
			if err := v.Stop(); err != nil {
				return err
			}
		}
		// stop configure manager
		for _, v := range a.configureManagerList {
			if err := v.Stop(); err != nil {
				return err
			}
		}
		return nil
	}
	return nil
}

func (a *Application) Stop() error {
	return a.stopFn()
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
	for _, name := range gen.OriginalGeneratorNames() {
		if _, ok := a.sourceConnectorGeneratorMap[name]; ok {
			return errors.New("source connector generator already exists:" + name)
		}
		a.sourceConnectorGeneratorMap[name] = gen
	}
	return nil
}

func (a *Application) AddTargetConnectorGenerator(gen pluginapi.TargetConnectorGenerator) error {
	for _, name := range gen.OriginalGeneratorNames() {
		if !strings.HasPrefix(name, "&") {
			name = "&" + name
		}
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
		fileManagerMap:   map[string]pluginapi.FileResourceManager{},
		configureManager: NewNestedConfigureManager(),

		sourceConnectorGeneratorMap:            map[string]pluginapi.SourceConnectorGenerator{},
		targetConnectorGeneratorMap:            map[string]pluginapi.TargetConnectorGenerator{},
		preInitializedSourceConnectorGenerator: map[string]pluginapi.SourceConnectorGenerator{},
		preInitializedTargetConnectorGenerator: map[string]pluginapi.TargetConnectorGenerator{},
	}
}
