package modelinst

type mapWrapper struct {
	m map[string]interface{}
}

func (m mapWrapper) AddOrUpdateField0(path []string, value interface{}) error {
	//TODO implement me
	panic("implement me")
}

func (m mapWrapper) GetFieldUnsafe0(path []string) interface{} {
	//TODO implement me
	panic("implement me")
}

func (m mapWrapper) RemoveObjectByPath(paths []string) error {
	//TODO implement me
	panic("implement me")
}

func (m mapWrapper) transferPrimitiveArray(srcName, dstName string, dst ModelInst2) error {
	//TODO implement me
	panic("implement me")
}

func (m mapWrapper) transferValue(srcName, dstName string, dst ModelInst2) error {
	//TODO implement me
	panic("implement me")
}

func (m mapWrapper) ensureSubObject(name string) (ModelInst2, error) {
	//TODO implement me
	panic("implement me")
}

func (m mapWrapper) getSubObject(name string) (ModelInst2, error) {
	//TODO implement me
	panic("implement me")
}

func (m mapWrapper) ensureSubArrayWithObjectElem(name string) (ModelInst2, error) {
	//TODO implement me
	panic("implement me")
}

func (m mapWrapper) getSubArrayWithObjectElem(name string) (ModelInst2, error) {
	//TODO implement me
	panic("implement me")
}

func (m mapWrapper) foreachArrayElement(f func(inst2 ModelInst2) error) error {
	//TODO implement me
	panic("implement me")
}

func (m mapWrapper) ensureArrayElement() (ModelInst2, error) {
	//TODO implement me
	panic("implement me")
}

func (m mapWrapper) ensureArrayElementWithIndex(idx int) (ModelInst2, error) {
	//TODO implement me
	panic("implement me")
}

func (m mapWrapper) putPrimitiveArray(name string, arr []interface{}) error {
	//TODO implement me
	panic("implement me")
}

func (m mapWrapper) setPrimitiveArrayIndex(name string, index int, value interface{}) error {
	//TODO implement me
	panic("implement me")
}

func (m mapWrapper) putPrimitiveValue(name string, val interface{}) error {
	//TODO implement me
	panic("implement me")
}
