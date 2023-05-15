package basicapi

type BasicContainer interface {
	RegisterCustomFn(name string, fnGen FnGen) error

	LoadFlowModel(tomlContent string) error
	LoadMerged(content string) error

	StartContainer() error
}
