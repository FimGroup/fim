package pluginapi

type SourceConnectorGenerateRequest struct {
	CommonSourceConnectorGenerateRequest
	InstanceName string
	Definition   *MappingDefinition
	Container    Container
}

type CommonSourceConnectorGenerateRequest struct {
	Options     map[string]string
	Application ApplicationSupport
}

type SourceConnectorGenerator interface {
	OriginalGeneratorNames() []string
	GenerateSourceConnectorInstance(req SourceConnectorGenerateRequest) (SourceConnector, error)

	InitializeSubGeneratorInstance(req CommonSourceConnectorGenerateRequest) (SourceConnectorGenerator, error)
	Startup() error
	Stop() error
}

type TargetConnectorGenerateRequest struct {
	CommonTargetConnectorGenerateRequest
	InstanceName string
	Definition   *MappingDefinition
	Container    Container
}

type CommonTargetConnectorGenerateRequest struct {
	Options     map[string]string
	Application ApplicationSupport
}

type TargetConnectorGenerator interface {
	OriginalGeneratorNames() []string
	GenerateTargetConnectorInstance(req TargetConnectorGenerateRequest) (TargetConnector, error)

	InitializeSubGeneratorInstance(req CommonTargetConnectorGenerateRequest) (TargetConnectorGenerator, error)
	Startup() error
	Stop() error
}
