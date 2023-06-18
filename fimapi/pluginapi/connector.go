package pluginapi

type SourceConnectorGenerateRequest struct {
	Options map[string]string

	Container   Container
	Application ApplicationSupport
}

type SourceConnectorGenerator interface {
	GeneratorNames() []string
	GenerateSourceConnectorInstance(req SourceConnectorGenerateRequest) (SourceConnector, error)
}

type TargetConnectorGenerateRequest struct {
	Options    map[string]string
	Definition *MappingDefinition

	Container   Container
	Application ApplicationSupport
}

type TargetConnectorGenerator interface {
	GeneratorNames() []string
	GenerateTargetConnectorInstance(req TargetConnectorGenerateRequest) (TargetConnector, error)
}
