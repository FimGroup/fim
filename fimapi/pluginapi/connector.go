package pluginapi

type SourceConnectorGenerator interface {
	GeneratorName() string
	GenerateSourceConnectorInstance(options map[string]string, container Container) (*struct {
		Connector
		ConnectorProcessEntryPoint
		InstanceName string
	}, error)
}

type TargetConnectorGenerator interface {
	GeneratorName() string
	GenerateTargetConnectorInstance(options map[string]string, container Container, definition *MappingDefinition) (*struct {
		Connector
		ConnectorFlow
		InstanceName string
	}, error)
}
