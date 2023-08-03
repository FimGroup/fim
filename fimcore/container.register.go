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

func newContainer(application *Application, businessName string) *ContainerInst {
	var flowModelMap = NewDataTypeDefinitions()

	return &ContainerInst{
		businessName: businessName,

		flowModelRawContents: [][]byte{},
		flowRawMap:           map[string]struct{ tf *templateFlow }{},
		flowMap:              map[string]*Flow{},
		flowModel:            flowModelMap,

		pipelineMap:        map[string]*Pipeline{},
		pipelineRawContent: map[string]struct{ *Pipeline }{},

		builtinGenFnMap: map[string]pluginapi.FnGen{},
		customGenFnMap:  map[string]pluginapi.FnGen{},

		connectorMap: map[string]pluginapi.Connector{},

		configureManager: NewNestedConfigureManager(),

		_logger:        loggerManager.GetLogger("FimCore.Container"),
		_loggerManager: loggerManager,
		application:    application,
	}
}

type ContainerInst struct {
	businessName    string
	dispatchDecider pluginapi.DispatchDecider

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

	connectorMap map[string]pluginapi.Connector

	lifecycleListeners []pluginapi.LifecycleListener

	configureManager *NestedConfigureManager

	_logger        providers.Logger
	_loggerManager providers.LoggerManager
	application    *Application

	stopFunction func() error
}

func (c *ContainerInst) SetupDispatchDecider(decider pluginapi.DispatchDecider) error {
	c.dispatchDecider = decider
	return nil
}

func (c *ContainerInst) AddLifecycleListener(listener pluginapi.LifecycleListener) {
	c.lifecycleListeners = append(c.lifecycleListeners, listener)
}

func (c *ContainerInst) StopContainer() error {
	return c.stopFunction()
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

type containerDispatchDeciderLifecycleListener struct {
	c *ContainerInst
}

func (c containerDispatchDeciderLifecycleListener) OnStart() error {
	return c.c.dispatchDecider.StartDispatcher()
}

func (c containerDispatchDeciderLifecycleListener) OnStop() error {
	return c.c.dispatchDecider.StopDispatcher()
}

func generateDispatchDeciderLifecycleListener(c *ContainerInst) pluginapi.LifecycleListener {
	return containerDispatchDeciderLifecycleListener{c}
}

func (c *ContainerInst) StartContainer() error {
	// internal mechanism registration
	c.AddLifecycleListener(generateDispatchDeciderLifecycleListener(c))

	// setup pipelines
	for _, p := range c.pipelineMap {
		if err := p.combinePipelineAndSourceConnector(); err != nil {
			return err
		}
	}

	allInit := false
	// cleanup unfinished initializations to avoid resource leak
	defer func() {
		if !allInit {
			for _, c := range c.connectorMap {
				if err := c.Stop(); err != nil {
					// omit error
				}
			}
		}
	}()

	// start connectors to accept requests
	for _, c := range c.connectorMap {
		if err := c.Start(); err != nil {
			return err
		}
	}
	allInit = true

	// trigger lifecycle listeners at end
	for _, v := range c.lifecycleListeners {
		if err := v.OnStart(); err != nil {
			return err
		}
	}

	c._logger.Info("start container success!")

	c.stopFunction = func() error {
		// trigger lifecycle listeners at start
		for _, v := range c.lifecycleListeners {
			if err := v.OnStop(); err != nil {
				return err
			}
		}

		// stop connectors
		for _, v := range c.connectorMap {
			if err := v.Stop(); err != nil {
				return err
			}
		}

		return nil
	}

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
