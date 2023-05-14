package fimcore

import (
	"errors"
	"fmt"
	"log"

	"github.com/ThisIsSun/fim/fimapi/basicapi"
	"github.com/ThisIsSun/fim/fimapi/pluginapi"
)

func NewUseContainer() basicapi.BasicContainer {
	return NewContainer()
}

func NewContainer() *ContainerInst {
	var flowModelMap = NewDataTypeDefinitions()

	return &ContainerInst{
		flowModelRawContents: [][]byte{},
		flowRawMap:           map[string]struct{ FlowTomlRaw []byte }{},
		flowMap:              map[string]*Flow{},
		flowModel:            flowModelMap,

		pipelineMap:        map[string]*Pipeline{},
		pipelineRawContent: map[string]struct{ PipelineTomlRaw []byte }{},

		builtinGenFnMap: map[string]pluginapi.FnGen{},
		customGenFnMap:  map[string]pluginapi.FnGen{},

		connectorMap:               map[string]pluginapi.Connector{},
		registerSourceConnectorGen: map[string]pluginapi.SourceConnectorGenerator{},
		registerTargetConnectorGen: map[string]pluginapi.TargetConnectorGenerator{},
	}
}

type ContainerInst struct {
	flowRawMap map[string]struct {
		FlowTomlRaw []byte
	}
	flowMap map[string]*Flow

	flowModelRawContents [][]byte
	flowModel            *DataTypeDefinitions

	pipelineRawContent map[string]struct {
		PipelineTomlRaw []byte
	}
	pipelineMap map[string]*Pipeline

	builtinGenFnMap map[string]pluginapi.FnGen
	customGenFnMap  map[string]pluginapi.FnGen

	connectorMap               map[string]pluginapi.Connector
	registerSourceConnectorGen map[string]pluginapi.SourceConnectorGenerator
	registerTargetConnectorGen map[string]pluginapi.TargetConnectorGenerator
}

func (c *ContainerInst) LoadFlowModel(tomlContent string) error {

	if err := c.flowModel.MergeToml(tomlContent); err != nil {
		return err
	}
	c.flowModelRawContents = append(c.flowModelRawContents, []byte(tomlContent))

	return nil
}

func (c *ContainerInst) LoadFlow(flowName, tomlContent string) error {
	_, ok := c.flowMap[flowName]
	if ok {
		return errors.New(fmt.Sprint("flow exists:", flowName))
	}

	flow := NewFlow(c.flowModel, c)
	if err := flow.MergeToml(tomlContent); err != nil {
		return err
	}
	c.flowMap[flowName] = flow
	c.flowRawMap[flowName] = struct{ FlowTomlRaw []byte }{
		FlowTomlRaw: []byte(tomlContent),
	}

	return nil
}

func (c *ContainerInst) LoadPipeline(pipelineName, tomlContent string) error {
	_, ok := c.pipelineMap[pipelineName]
	if ok {
		return errors.New(fmt.Sprintf("pipeline exists:%s", pipelineName))
	}

	p, err := NewPipeline(tomlContent, c)
	if err != nil {
		return err
	}
	c.pipelineRawContent[pipelineName] = struct{ PipelineTomlRaw []byte }{PipelineTomlRaw: []byte(tomlContent)}
	c.pipelineMap[pipelineName] = p

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
