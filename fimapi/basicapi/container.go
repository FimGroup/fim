package basicapi

type BasicContainer interface {
	RegisterCustomFn(name string, fnGen FnGen) error
	AddConfigureManager(manager ConfigureManager) error

	LoadFlowModel(tomlContent string) error
	LoadMerged(content string) error

	StartContainer() error
}
