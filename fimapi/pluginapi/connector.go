package pluginapi

type SourceConnectorGenerator interface {
	GeneratorNames() []string
	GenerateSourceConnectorInstance(options map[string]string, container Container) (*struct {
		Connector
		ConnectorProcessEntryPoint
	}, error)
}

type TargetConnectorGenerator interface {
	GeneratorNames() []string
	GenerateTargetConnectorInstance(options map[string]string, container Container, definition *MappingDefinition) (*struct {
		Connector
		ConnectorFlow
	}, error)
}
