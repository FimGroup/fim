package modelinst

import "errors"

var _ ModelInst2 = defaultModelInst2{}

type defaultModelInst2 struct {
}

func (d defaultModelInst2) transferPrimitiveArray(srcName, dstName string, dst ModelInst2) error {
	return errors.New("operation unsupported")
}

func (d defaultModelInst2) transferValue(srcName, dstName string, dst ModelInst2) error {
	return errors.New("operation unsupported")
}

func (d defaultModelInst2) ensureSubObject(name string) (ModelInst2, error) {
	return nil, errors.New("operation unsupported")
}

func (d defaultModelInst2) getSubObject(name string) (ModelInst2, error) {
	return nil, errors.New("operation unsupported")
}

func (d defaultModelInst2) ensureSubArrayWithObjectElem(name string) (ModelInst2, error) {
	return nil, errors.New("operation unsupported")
}

func (d defaultModelInst2) getSubArrayWithObjectElem(name string) (ModelInst2, error) {
	return nil, errors.New("operation unsupported")
}

func (d defaultModelInst2) foreachArrayElement(f func(inst2 ModelInst2) error) error {
	return errors.New("operation unsupported")
}

func (d defaultModelInst2) ensureArrayElement() (ModelInst2, error) {
	return nil, errors.New("operation unsupported")
}

func (d defaultModelInst2) ensureArrayElementWithIndex(idx int) (ModelInst2, error) {
	return nil, errors.New("operation unsupported")
}

func (d defaultModelInst2) putPrimitiveArray(name string, arr []interface{}) error {
	return errors.New("operation unsupported")
}

func (d defaultModelInst2) setPrimitiveArrayIndex(name string, index int, value interface{}) error {
	return errors.New("operation unsupported")
}

func (d defaultModelInst2) putPrimitiveValue(name string, val interface{}) error {
	return errors.New("operation unsupported")
}

func (d defaultModelInst2) RemoveObjectByPath(paths []string) error {
	return errors.New("operation unsupported")
}

func (d defaultModelInst2) AddOrUpdateField0(path []string, value interface{}) error {
	return errors.New("operation unsupported")
}

func (d defaultModelInst2) GetFieldUnsafe0(path []string) interface{} {
	//FIXME should break flow?
	return nil
}

func (d defaultModelInst2) ToGeneralObject() interface{} {
	//FIXME should break flow?
	return nil
}
