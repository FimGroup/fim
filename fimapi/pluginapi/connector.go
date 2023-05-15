package pluginapi

type SourceConnectorGenerator interface {
	GeneratorNames() []string
	GenerateSourceConnectorInstance(options map[string]string, container Container) (SourceConnector, error)
}

type TargetConnectorGenerator interface {
	GeneratorNames() []string
	GenerateTargetConnectorInstance(options map[string]string, container Container, definition *MappingDefinition) (TargetConnector, error)
}
