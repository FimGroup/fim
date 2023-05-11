package esbcore

import (
	"errors"
	"fmt"

	"esbconcept/esbapi"
)

func NewContainer() *ContainerInst {
	flowRawMap := map[string]struct{ FlowTomlRaw []byte }{}
	flowMap := map[string]*Flow{}

	var flowModelRawMap [][]byte
	var flowModelMap = NewDataTypeDefinitions()

	return &ContainerInst{
		flowModelRawContents: flowModelRawMap,
		flowRawMap:           flowRawMap,
		flowMap:              flowMap,
		flowModel:            flowModelMap,

		pipelineMap:        map[string]*Pipeline{},
		pipelineRawContent: map[string]struct{ PipelineTomlRaw []byte }{},

		builtinGenFnMap: map[string]esbapi.FnGen{},
		customGenFnMap:  map[string]esbapi.FnGen{},
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

	builtinGenFnMap map[string]esbapi.FnGen
	customGenFnMap  map[string]esbapi.FnGen
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

func (c *ContainerInst) RunPipelines() {
	for _, p := range c.pipelineMap {
		go func() {
			if err := p.RunPipeline(); err != nil {
				panic(err)
			}
		}()
	}
}
