package fimcore

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/FimGroup/fim/fimapi/pluginapi"

	"github.com/pelletier/go-toml/v2"
)

type connectorGeneratorDefinition struct {
	SourceConnectorDefMapping map[string]map[string]string `toml:"source_connector"`
	TargetConnectorDefMapping map[string]map[string]string `toml:"target_connector"`
}

func (a *Application) internalGenerateSourceConnectorInstance(name, instanceName string, container pluginapi.Container, options map[string]string, mappingDef *pluginapi.MappingDefinition) (pluginapi.SourceConnector, error) {
	req := pluginapi.SourceConnectorGenerateRequest{
		CommonSourceConnectorGenerateRequest: pluginapi.CommonSourceConnectorGenerateRequest{
			Options:     options,
			Application: a,
		},
		InstanceName: instanceName,
		Definition:   mappingDef,
		Container:    container,
	}
	if gen, ok := a.preInitializedSourceConnectorGenerator[name]; ok {
		return gen.GenerateSourceConnectorInstance(req)
	} else if gen, ok := a.sourceConnectorGeneratorMap[name]; ok {
		return gen.GenerateSourceConnectorInstance(req)
	} else {
		return nil, errors.New("unknown source connector generator:" + name)
	}
}

func (a *Application) internalGenerateTargetConnectorInstance(name, instanceName string, container pluginapi.Container, options map[string]string, mappingDef *pluginapi.MappingDefinition) (pluginapi.TargetConnector, error) {
	req := pluginapi.TargetConnectorGenerateRequest{
		CommonTargetConnectorGenerateRequest: pluginapi.CommonTargetConnectorGenerateRequest{
			Options:     options,
			Application: a,
		},
		InstanceName: instanceName,
		Definition:   mappingDef,
		Container:    container,
	}
	if gen, ok := a.preInitializedTargetConnectorGenerator[name]; ok {
		return gen.GenerateTargetConnectorInstance(req)
	} else if gen, ok := a.targetConnectorGeneratorMap[name]; ok {
		return gen.GenerateTargetConnectorInstance(req)
	} else {
		return nil, errors.New("unknown target connector generator:" + name)
	}
}

func (a *Application) AddSubConnectorGeneratorDefinitions(tomlData string) error {
	def := new(connectorGeneratorDefinition)
	if err := toml.NewDecoder(bytes.NewBufferString(tomlData)).DisallowUnknownFields().Decode(def); err != nil {
		return err
	}
	a.sourceConnectorGeneratorDefinitions = append(a.sourceConnectorGeneratorDefinitions, def.SourceConnectorDefMapping)
	a.targetConnectorGeneratorDefinitions = append(a.targetConnectorGeneratorDefinitions, def.TargetConnectorDefMapping)
	return nil
}

func (a *Application) setupAndStoreSubSourceConnectorGenerator(def map[string]map[string]string) error {
	for name, options := range def {
		if newOptions, err := a.dealWithConfigurableOptions(options); err != nil {
			return err
		} else {
			options = newOptions
		}
		parent, ok := options["@parent"]
		if !ok {
			return errors.New("unknown parent of creating source connector generator:" + name)
		}
		gen, ok := a.sourceConnectorGeneratorMap[parent]
		if !ok {
			return errors.New(fmt.Sprintf("no source connector generator=%s found for %s", parent, name))
		}
		subGen, err := gen.InitializeSubGeneratorInstance(pluginapi.CommonSourceConnectorGenerateRequest{
			Options:     options,
			Application: a,
		})
		if err != nil {
			return err
		} else {
			a.preInitializedSourceConnectorGenerator[parent] = subGen
		}
	}
	return nil
}

func (a *Application) setupAndStoreSubTargetConnectorGenerator(def map[string]map[string]string) error {
	for name, options := range def {
		if newOptions, err := a.dealWithConfigurableOptions(options); err != nil {
			return err
		} else {
			options = newOptions
		}
		parent, ok := options["@parent"]
		if !ok {
			return errors.New("unknown parent of creating target connector generator:" + name)
		}
		gen, ok := a.targetConnectorGeneratorMap[parent]
		if !ok {
			return errors.New(fmt.Sprintf("no target connector generator=%s found for %s", parent, name))
		}
		subGen, err := gen.InitializeSubGeneratorInstance(pluginapi.CommonTargetConnectorGenerateRequest{
			Options:     options,
			Application: a,
		})
		if err != nil {
			return err
		} else {
			a.preInitializedTargetConnectorGenerator[parent] = subGen
		}
	}
	return nil
}

func (a *Application) dealWithConfigurableOptions(options map[string]string) (map[string]string, error) {
	newOptions := map[string]string{}
	for k, v := range options {
		newOptions[k] = a.configureManager.ReplaceStaticConfigure(v)
	}
	return newOptions, nil
}
