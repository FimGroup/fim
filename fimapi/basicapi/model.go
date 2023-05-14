package basicapi

type Model interface {
	//FIXME need design the type and behaviors of Model
	AddOrUpdateField0(paths []string, value interface{}) error
	GetFieldUnsafe(paths []string) interface{}
}
