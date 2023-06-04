package basicapi

type Model interface {
	//FIXME need design the type and behaviors of Model

	//AddOrUpdateField0 supports primitive value only from object and array(nested array is ok)
	AddOrUpdateField0(path []string, value interface{}) error
	//GetFieldUnsafe0 supports primitive value only from object and array(nested array is ok)
	GetFieldUnsafe0(path []string) interface{}
}
