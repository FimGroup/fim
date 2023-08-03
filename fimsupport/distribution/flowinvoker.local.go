package distribution

import (
	"errors"

	"github.com/FimGroup/fim/fimapi/pluginapi"
)

type LocalFlowInvoker struct {
	flowMapping map[string]pluginapi.PipelineProcess
}

func (l *LocalFlowInvoker) AddPipeline(pipelineName string, process pluginapi.PipelineProcess) error {
	_, ok := l.flowMapping[pipelineName]
	if ok {
		return errors.New("pipeline already exists:" + pipelineName)
	}
	l.flowMapping[pipelineName] = process
	return nil
}

func (l *LocalFlowInvoker) Metadata() pluginapi.FlowInvokerMeta {
	return pluginapi.NewFlowInvokerMeta("local", false)
}

func (l *LocalFlowInvoker) Invoke(pipelineFullName string, model pluginapi.Model) error {
	flow, ok := l.flowMapping[pipelineFullName]
	if !ok {
		return errors.New("no pipeline found:" + pipelineFullName)
	}
	return flow(model)
}

func (l *LocalFlowInvoker) StartFlowInvoker() error {
	return nil
}

func (l *LocalFlowInvoker) StopFlowInvoker() error {
	return nil
}

func NewLocalFlowInvoker() pluginapi.FlowInvoker {
	return &LocalFlowInvoker{
		flowMapping: map[string]pluginapi.PipelineProcess{},
	}
}
