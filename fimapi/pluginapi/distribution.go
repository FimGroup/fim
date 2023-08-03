package pluginapi

import (
	"fmt"
)

func ConcatFullPipelineName(serviceName, pipelineName string) string {
	return fmt.Sprintf("%s/%s", serviceName, pipelineName)
}

type FlowInvokerMeta struct {
	name   string
	remote bool
}

func NewFlowInvokerMeta(name string, remote bool) FlowInvokerMeta {
	return FlowInvokerMeta{
		name:   name,
		remote: remote,
	}
}

func (m FlowInvokerMeta) Remote() bool {
	return m.remote
}

func (m FlowInvokerMeta) Name() string {
	return m.name
}

type DispatchDecider interface {
	AddFlowInvoker(f FlowInvoker) error

	InjectLocalPipeline(pipelineFullName string, process PipelineProcess) error
	PipelineDispatcher(pipelineFullName string) PipelineProcess

	StartDispatcher() error // triggered once all pipeline resources prepared
	StopDispatcher() error  // triggered once all pipeline resources prepared
}

type FlowInvoker interface {
	Metadata() FlowInvokerMeta

	AddPipeline(pipelineName string, process PipelineProcess) error
	Invoke(pipelineFullName string, model Model) error

	StartFlowInvoker() error // controlled by DispatcherDecider
	StopFlowInvoker() error  // controlled by DispatcherDecider
}
