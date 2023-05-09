package esbcore

import (
	"errors"
	"fmt"
)

func NewContainer() *Container {
	flowRawMap := map[string]struct{ FlowTomlRaw []byte }{}
	flowMap := map[string]*Flow{}

	var flowModelRawMap [][]byte
	var flowModelMap = NewDataTypeDefinitions()

	return &Container{
		flowModelRawContents: flowModelRawMap,
		flowRawMap:           flowRawMap,
		flowMap:              flowMap,
		flowModel:            flowModelMap,
	}
}

type Container struct {
	flowRawMap map[string]struct {
		FlowTomlRaw []byte
	}
	flowMap map[string]*Flow

	flowModelRawContents [][]byte
	flowModel            *DataTypeDefinitions
}

func (c *Container) LoadFlowModel(tomlContent string) error {

	if err := c.flowModel.MergeToml(tomlContent); err != nil {
		return err
	}
	c.flowModelRawContents = append(c.flowModelRawContents, []byte(tomlContent))

	return nil
}

func (c *Container) LoadFlow(flowName, tomlContent string) error {
	_, ok := c.flowMap[flowName]
	if ok {
		return errors.New(fmt.Sprint("flow exists:", flowName))
	}

	flow := NewFlow(c.flowModel)
	if err := flow.MergeToml(tomlContent); err != nil {
		return err
	}
	c.flowMap[flowName] = flow
	c.flowRawMap[flowName] = struct{ FlowTomlRaw []byte }{
		FlowTomlRaw: []byte(tomlContent),
	}

	return nil
}
