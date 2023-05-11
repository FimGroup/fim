package esbapi

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

type Model interface {
	//FIXME need design the type and behaviors of Model
	AddOrUpdateField0(paths []string, value interface{}) error
	GetFieldUnsafe(paths []string) interface{}
}

type Container interface {
	RegisterBuiltinFn(name string, fnGen FnGen) error
	RegisterCustomFn(name string, fnGen FnGen) error
}
