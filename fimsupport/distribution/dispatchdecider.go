package distribution

import (
	"errors"

	"github.com/FimGroup/fim/fimapi/pluginapi"
)

// AbstractDispatchDecider is used for dispatching request to next step
// A decider should make decisions based on the considerations of:
// * Interaction types like rpc/event
// * Keep minimum request/response serialization cost
// * Burden of communication or latency
// * Other resource concerns
// When enabling distributed processing model, multiple nodes will be working as:
// * Dispatcher mode: entrypoint node(usually the source connector node) dispatches each stage of the pipeline to different nodes
// * Relay/SEDA mode: the request will be transferred one by one node in the pipeline and finally return to the entrypoint node to respond result(if required)
// * Note: choosing mode should have the following considerations:
//   - passing full context
//   - communication cost
//   - require more on serialization
//   - other resource concerns
type AbstractDispatchDecider struct {
}

type SingleDispatchDecider struct {
	AbstractDispatchDecider
	flowInvoker pluginapi.FlowInvoker
}

func (s *SingleDispatchDecider) AddFlowInvoker(f pluginapi.FlowInvoker) error {
	if s.flowInvoker != nil {
		return errors.New("FlowInvoker already added")
	}

	s.flowInvoker = f
	return nil
}

func (s *SingleDispatchDecider) InjectLocalPipeline(pipelineFullName string, process pluginapi.PipelineProcess) error {
	return s.flowInvoker.AddPipeline(pipelineFullName, process)
}

func (s *SingleDispatchDecider) PipelineDispatcher(pipelineFullName string) pluginapi.PipelineProcess {
	return func(m pluginapi.Model) error {
		return s.flowInvoker.Invoke(pipelineFullName, m)
	}
}

func (s *SingleDispatchDecider) StartDispatcher() error {
	return s.flowInvoker.StartFlowInvoker()
}

func (s *SingleDispatchDecider) StopDispatcher() error {
	return s.flowInvoker.StopFlowInvoker()
}

func NewSingleDispatchDecider() pluginapi.DispatchDecider {
	return new(SingleDispatchDecider)
}
