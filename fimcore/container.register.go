package fimcore

import (
	"log"

	"github.com/FimGroup/fim/fimapi/basicapi"
	"github.com/FimGroup/fim/fimapi/pluginapi"
)

func NewUseContainer() basicapi.BasicContainer {
	return NewContainer()
}

func NewContainer() *ContainerInst {
	var flowModelMap = NewDataTypeDefinitions()

	return &ContainerInst{
		flowModelRawContents: [][]byte{},
		flowRawMap:           map[string]struct{ tf *templateFlow }{},
		flowMap:              map[string]*Flow{},
		flowModel:            flowModelMap,

		pipelineMap:        map[string]*Pipeline{},
		pipelineRawContent: map[string]struct{ *Pipeline }{},

		builtinGenFnMap: map[string]pluginapi.FnGen{},
		customGenFnMap:  map[string]pluginapi.FnGen{},

		connectorMap:               map[string]pluginapi.Connector{},
		registerSourceConnectorGen: map[string]pluginapi.SourceConnectorGenerator{},
		registerTargetConnectorGen: map[string]pluginapi.TargetConnectorGenerator{},

		configureManager: NewNestedConfigureManager(),
	}
}

type ContainerInst struct {
	flowRawMap map[string]struct {
		tf *templateFlow
	}
	flowMap map[string]*Flow

	flowModelRawContents [][]byte
	flowModel            *DataTypeDefinitions

	pipelineRawContent map[string]struct {
		*Pipeline
	}
	pipelineMap map[string]*Pipeline

	builtinGenFnMap map[string]pluginapi.FnGen
	customGenFnMap  map[string]pluginapi.FnGen

	connectorMap               map[string]pluginapi.Connector
	registerSourceConnectorGen map[string]pluginapi.SourceConnectorGenerator
	registerTargetConnectorGen map[string]pluginapi.TargetConnectorGenerator

	configureManager *NestedConfigureManager
}

func (c *ContainerInst) LoadFlowModel(tomlContent string) error {

	if err := c.flowModel.MergeToml(tomlContent); err != nil {
		return err
	}
	c.flowModelRawContents = append(c.flowModelRawContents, []byte(tomlContent))

	return nil
}

func (c *ContainerInst) StartContainer() error {
	// setup pipelines
	for _, p := range c.pipelineMap {
		if err := p.setupPipeline(); err != nil {
			return err
		}
	}

	allInit := false
	defer func() {
		if !allInit {
			for _, c := range c.connectorMap {
				if err := c.Stop(); err != nil {
					// omit error
				}
			}
		}
	}()
	// last step: start connectors to accept requests
	for _, c := range c.connectorMap {
		if err := c.Start(); err != nil {
			return err
		}
	}
	allInit = true

	log.Println("start container success!")
	return nil
}

func (c *ContainerInst) NewModel() pluginapi.Model {
	return NewModelInst(c.flowModel)
}

func (c *ContainerInst) AddConfigureManager(manager basicapi.ConfigureManager) error {
	c.configureManager.addSubConfigureManager(manager)
	return nil
}
