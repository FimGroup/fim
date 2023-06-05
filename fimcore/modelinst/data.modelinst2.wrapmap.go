package modelinst

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/FimGroup/fim/fimapi/rule"
)

type readonlyMapWrapper struct {
	m map[string]interface{}

	defaultModelInst2
}

func (m readonlyMapWrapper) ToGeneralObject() interface{} {
	return m.m
}

func (m readonlyMapWrapper) GetFieldUnsafe0(paths []string) interface{} {
	var parent interface{} = m.m
	for _, path := range paths[:len(paths)-1] {
		name, idx := rule.ExtractArrayPath(path)
		if idx < 0 {
			// handling object
			obj, ok := parent.(map[string]interface{})
			if !ok {
				return nil //FIXME should raise error?
			}
			elem, ok := obj[name]
			if !ok {
				return nil
			}
			parent = elem
			continue
		} else {
			// handling array
			arr, ok := parent.([]interface{})
			if !ok {
				return nil //FIXME should raise error?
			}
			if len(arr) > idx {
				parent = arr[idx]
				continue
			} else {
				return nil
			}
		}
	}
	// last level
	name, idx := rule.ExtractArrayPath(paths[len(paths)-1])
	if idx < 0 {
		// handling object
		obj, ok := parent.(map[string]interface{})
		if !ok {
			return nil //FIXME should raise error?
		}
		elem, ok := obj[name]
		if !ok {
			return nil
		}
		if isPrimitive(elem) {
			return elem
		} else {
			return nil //FIXME should raise error?
		}
	} else {
		// handling array
		arr, ok := parent.([]interface{})
		if !ok {
			return nil //FIXME should raise error?
		}
		if len(arr) > idx {
			elem := arr[idx]
			if isPrimitive(elem) {
				return elem
			} else {
				return nil //FIXME should raise error?
			}
		} else {
			return nil
		}
	}
}

func (m readonlyMapWrapper) transferPrimitiveArray(srcName, dstName string, dst ModelInst2) error {
	arrObj, ok := m.m[srcName]
	if !ok {
		return nil
	}
	arr, ok := arrObj.([]interface{})
	if !ok {
		return errors.New("src object is not array, fieldName:" + srcName)
	}
	for _, v := range arr {
		if v == nil {
			continue
		}
		if !isPrimitive(v) {
			return errors.New("try to  transfer primitive array on an non-primitive array, fieldName:" + srcName)
		} else {
			break
		}
	}
	return dst.putPrimitiveArray(dstName, arr)
}

func (m readonlyMapWrapper) transferValue(srcName, dstName string, dst ModelInst2) error {
	if val, ok := m.m[srcName]; !ok {
		return nil
	} else if !isPrimitive(val) {
		return errors.New("try to transfer non-primitive value, fieldName:" + srcName)
	} else {
		return dst.putPrimitiveValue(dstName, val)
	}
}

func (m readonlyMapWrapper) getSubObject(name string) (ModelInst2, error) {
	val, ok := m.m[name]
	if !ok {
		return nil, nil
	}
	if val == nil {
		return nil, nil
	}
	switch tv := val.(type) {
	case map[string]interface{}:
		return readonlyMapWrapper{m: tv}, nil
	default:
		return nil, errors.New(fmt.Sprintf("field=[%s] type=[%s] is not object", name, reflect.TypeOf(tv)))
	}
}

func (m readonlyMapWrapper) getSubArrayWithObjectElem(name string) (ModelInst2, error) {
	val, ok := m.m[name]
	if !ok {
		return nil, nil
	}
	if val == nil {
		return nil, nil
	}
	switch tv := val.(type) {
	case []interface{}:
		for _, v := range tv {
			if v == nil {
				continue
			} else if _, ok := v.(map[string]interface{}); ok {
				return readonlyArrayWrapper{data: tv}, nil
			} else {
				return nil, errors.New(fmt.Sprintf("field=[%s] type=[%s] is not object array or the element is not object", name, reflect.TypeOf(tv)))
			}
		}
		// default, regarded as object array
		return readonlyArrayWrapper{data: tv}, nil
	default:
		return nil, errors.New(fmt.Sprintf("field=[%s] type=[%s] is not object array", name, reflect.TypeOf(tv)))
	}
}

func (m readonlyMapWrapper) foreachArrayElement(f func(inst2 ModelInst2) error) error {
	return errors.New("foreachArrayElement is not supported by object type")
}

type readonlyArrayWrapper struct {
	data []interface{}

	defaultModelInst2
}

func (m readonlyArrayWrapper) ToGeneralObject() interface{} {
	return m.data
}

func (m readonlyArrayWrapper) GetFieldUnsafe0(path []string) interface{} {
	return errors.New("GetFieldUnsafe0 is not supported by object array type")
}

func (m readonlyArrayWrapper) transferPrimitiveArray(srcName, dstName string, dst ModelInst2) error {
	return errors.New("transferPrimitiveArray is not supported by object array type")
}

func (m readonlyArrayWrapper) transferValue(srcName, dstName string, dst ModelInst2) error {
	return errors.New("transferValue is not supported by object array type")
}

func (m readonlyArrayWrapper) getSubObject(name string) (ModelInst2, error) {
	return nil, errors.New("getSubObject is not supported by object array type")
}

func (m readonlyArrayWrapper) getSubArrayWithObjectElem(name string) (ModelInst2, error) {
	return nil, errors.New("getSubArrayWithObjectElem is not supported by object array type")
}

func (m readonlyArrayWrapper) foreachArrayElement(f func(inst2 ModelInst2) error) error {
	for _, v := range m.data {
		inst, ok := v.(map[string]interface{})
		if !ok {
			return errors.New("foreachArrayElement has non-object element")
		}
		if err := f(readonlyMapWrapper{m: inst}); err != nil {
			return err
		}
	}
	return nil
}

type readonlyPrimitiveArrayWrapper struct {
	data []interface{}

	defaultModelInst2
}

func (m readonlyPrimitiveArrayWrapper) ToGeneralObject() interface{} {
	return m.data
}

func (m readonlyPrimitiveArrayWrapper) GetFieldUnsafe0(path []string) interface{} {
	return errors.New("GetFieldUnsafe0 is not supported by primitive array")
}

func (m readonlyPrimitiveArrayWrapper) transferPrimitiveArray(srcName, dstName string, dst ModelInst2) error {
	return errors.New("transferPrimitiveArray is not supported by primitive array")
}

func (m readonlyPrimitiveArrayWrapper) transferValue(srcName, dstName string, dst ModelInst2) error {
	return errors.New("transferValue is not supported by primitive array")
}

func (m readonlyPrimitiveArrayWrapper) getSubObject(name string) (ModelInst2, error) {
	return nil, errors.New("getSubObject is not supported by primitive array")
}

func (m readonlyPrimitiveArrayWrapper) getSubArrayWithObjectElem(name string) (ModelInst2, error) {
	return nil, errors.New("getSubArrayWithObjectElem is not supported by primitive array")
}

func (m readonlyPrimitiveArrayWrapper) foreachArrayElement(f func(inst2 ModelInst2) error) error {
	return errors.New("foreachArrayElement is not supported by primitive array")
}

type readonlyElementWrapper struct {
	data interface{}

	defaultModelInst2
}

func (m readonlyElementWrapper) ToGeneralObject() interface{} {
	return m.data
}

func (m readonlyElementWrapper) GetFieldUnsafe0(path []string) interface{} {
	return nil
}

func (m readonlyElementWrapper) transferPrimitiveArray(srcName, dstName string, dst ModelInst2) error {
	return errors.New("transferPrimitiveArray is not supported by primitive type")
}

func (m readonlyElementWrapper) transferValue(srcName, dstName string, dst ModelInst2) error {
	return errors.New("transferValue is not supported by primitive type")
}

func (m readonlyElementWrapper) getSubObject(name string) (ModelInst2, error) {
	return nil, errors.New("getSubObject is not supported by primitive type")
}

func (m readonlyElementWrapper) getSubArrayWithObjectElem(name string) (ModelInst2, error) {
	return nil, errors.New("getSubArrayWithObjectElem is not supported by primitive type")
}

func (m readonlyElementWrapper) foreachArrayElement(f func(inst2 ModelInst2) error) error {
	return errors.New("foreachArrayElement is not supported by primitive type")
}
