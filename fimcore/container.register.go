package fimcore

import (
	"github.com/FimGroup/fim/fimapi/basicapi"
	"github.com/FimGroup/fim/fimapi/pluginapi"
	"github.com/FimGroup/fim/fimapi/providers"
	"github.com/FimGroup/fim/fimcore/modelinst"
)

var _ providers.ContainerProvided = new(ContainerInst)
var _ pluginapi.Container = new(ContainerInst)
var _ basicapi.BasicContainer = new(ContainerInst)

func newContainer(application *Application) *ContainerInst {
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

		_logger:        loggerManager.GetLogger("FimCore.Container"),
		_loggerManager: loggerManager,
		application:    application,
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

	_logger        providers.Logger
	_loggerManager providers.LoggerManager
	application    *Application
}

func (c *ContainerInst) GetContainerLoggerManager() providers.LoggerManager {
	return c._loggerManager
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

	c._logger.Info("start container success!")
	return nil
}

func (c *ContainerInst) NewModel() pluginapi.Model {
	return modelinst.ModelInstHelper{}.NewInst()
}

func (c *ContainerInst) WrapReadonlyModelFromMap(m map[string]interface{}) (pluginapi.Model, error) {
	//FIXME need to make sure readonly
	return modelinst.ModelInstHelper{}.WrapReadonlyMap(m), nil
}

func (c *ContainerInst) AddConfigureManager(manager basicapi.ConfigureManager) error {
	c.configureManager.addSubConfigureManager(manager)
	return nil
}
