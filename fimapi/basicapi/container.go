package basicapi

type BasicContainer interface {
	RegisterCustomFn(name string, fnGen FnGen) error

	LoadFlowModel(tomlContent string) error
	LoadFlow(flowName, tomlContent string) error
	LoadPipeline(pipelineName, tomlContent string) error

	StartContainer() error
}
