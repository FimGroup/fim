package pluginapi

import (
	"github.com/FimGroup/fim/fimapi/basicapi"
)

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
type ModelCopy interface {
	Transfer(dst Model) error
}
type ModelEncoding interface {
	ToToml() ([]byte, error)
	FromToml([]byte) error
}

type Container interface {
	RegisterBuiltinFn(name string, fnGen FnGen) error
	RegisterCustomFn(name string, fnGen FnGen) error

	NewModel() Model
	WrapReadonlyModelFromMap(map[string]interface{}) (Model, error)

	LoadFlowModel(tomlContent string) error
	LoadMerged(content string) error

	SetupDispatchDecider(decider DispatchDecider) error //TODO Temp solution: container level, due to lifecycle management
	AddLifecycleListener(listener LifecycleListener)
	StartContainer() error
	StopContainer() error
}

type PipelineProcess func(m Model) error
type MappingDefinition struct {
	ErrSimple    []map[string]string
	ReqConverter func(src, dst Model) error
	ReqArgPaths  []string
	ResConverter func(src, dst Model) error
	ResArgPaths  []string
}

type Connector interface {
	Start() error
	Stop() error
	Reload() error
}

type SourceConnector interface {
	Connector

	BindPipeline(PipelineProcess) error
}

type TargetConnector interface {
	Connector

	InvokeFlow(s, d Model) error
}
