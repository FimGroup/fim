package pluginapi

import "github.com/ThisIsSun/fim/fimapi/basicapi"

type DataType int

const (
	DataTypeUnavailable DataType = 0
	DataTypeInt         DataType = 1
	DataTypeString      DataType = 2
	DataTypeBool        DataType = 3
	DataTypeFloat       DataType = 4
	DataTypeArray       DataType = 51
	DataTypeObject      DataType = 52
)

const (
	PathSeparator = "/"
)

type TypeOfNode int

const (
	TypeUnknown       TypeOfNode = 0
	TypeDataNode      TypeOfNode = 1
	TypeNsNode        TypeOfNode = 2
	TypeAttributeNode TypeOfNode = 3
)

type Model = basicapi.Model

type Container interface {
	RegisterBuiltinFn(name string, fnGen FnGen) error
	RegisterCustomFn(name string, fnGen FnGen) error
	RegisterSourceConnectorGen(connGen SourceConnectorGenerator) error
	RegisterTargetConnectorGen(connGen TargetConnectorGenerator) error

	NewModel() Model

	LoadFlowModel(tomlContent string) error
	LoadMerged(content string) error

	StartContainer() error
}

type PipelineProcess func(m Model) error
type MappingDefinition struct {
	Req       [][]string
	Res       [][]string
	ErrSimple []map[string]string
}

type Connector interface {
	Start() error
	Stop() error
	Reload() error

	ConnectorName() string
}

type SourceConnector interface {
	Connector

	InvokeProcess(PipelineProcess, *MappingDefinition) error
}

type TargetConnector interface {
	Connector

	InvokeFlow(s, d Model) error
}
