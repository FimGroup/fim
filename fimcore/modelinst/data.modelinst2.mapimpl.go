package modelinst

import (
	"errors"
	"fmt"

	"github.com/FimGroup/fim/fimapi/rule"
)

const (
	valueTypePrimitive      = 1
	valueTypeObject         = 2
	valueTypeArray          = 3
	valueTypePrimitiveArray = 4
)

type modelInst2MapImpl struct {
	data         map[string]*modelInst2MapImpl
	array        []*modelInst2MapImpl
	primitiveArr []interface{}
	value        interface{}
	valueType    byte
}

func (m *modelInst2MapImpl) ToGeneralObject() interface{} {
	switch m.valueType {
	case valueTypePrimitive:
		return m.value
	case valueTypePrimitiveArray:
		arr := make([]interface{}, len(m.primitiveArr))
		for i, v := range m.primitiveArr {
			arr[i] = v
		}
		return arr
	case valueTypeObject:
		r := map[string]interface{}{}
		for k, v := range m.data {
			r[k] = v.ToGeneralObject()
		}
		return r
	case valueTypeArray:
		rarr := make([]interface{}, len(m.array))
		for i, v := range m.array {
			rarr[i] = v.ToGeneralObject()
		}
		return rarr
	default:
		panic("unknown value type:" + fmt.Sprint(m.valueType))
	}
}

func (m *modelInst2MapImpl) AddOrUpdateField0(pathLvs []string, value interface{}) error {
	// skip on nil
	if value == nil {
		return nil
	}
	// make sure value is acceptable
	value = mustConvertPrimitive(value)

	parent := m
	for _, path := range pathLvs[:len(pathLvs)-1] {
		if m.valueType != valueTypeObject {
			return errors.New("type is not object")
		}
		name, idx := rule.ExtractArrayPath(path)
		if idx < 0 {
			// handling object field
			sub, ok := parent.data[name]
			if !ok {
				newSub, err := parent.ensureSubObject(path)
				if err != nil {
					return err
				}
				sub = newSub.(*modelInst2MapImpl)
			}
			parent = sub
			continue
		} else {
			// handling array access
			sub, ok := parent.data[name]
			if !ok {
				newSub, err := parent.ensureSubArrayWithObjectElem(name)
				if err != nil {
					return err
				}
				sub = newSub.(*modelInst2MapImpl)
			}
			elem, err := sub.ensureArrayElementWithIndex(idx)
			if err != nil {
				return err
			}
			parent = elem.(*modelInst2MapImpl)
		}
	}

	// last level
	if parent.valueType != valueTypeObject {
		return errors.New("type is not object")
	}
	lastPath := pathLvs[len(pathLvs)-1]
	name, idx := rule.ExtractArrayPath(lastPath)
	if idx < 0 {
		// handling object field
		if err := parent.putPrimitiveValue(name, value); err != nil {
			return err
		}
		return nil
	} else {
		// handling array access - primitive array
		// set primitive value with given index
		if err := parent.setPrimitiveArrayIndex(name, idx, value); err != nil {
			return err
		}
		return nil
	}
}

func (m *modelInst2MapImpl) GetFieldUnsafe0(pathLvs []string) interface{} {
	parent := m
	for pathIdx, path := range pathLvs {
		if m.valueType != valueTypeObject {
			return errors.New("type is not object")
		}
		name, idx := rule.ExtractArrayPath(path)
		if idx < 0 {
			// handling object field
			sub, ok := parent.data[name]
			if !ok {
				return nil
			}
			parent = sub
			continue
		} else {
			// handling array access
			sub, ok := parent.data[name]
			if !ok {
				return nil
			}
			if sub.valueType == valueTypeArray {
				// object array
				if idx < len(sub.array) {
					parent = sub.array[idx]
					continue
				} else {
					return nil
				}
			} else if sub.valueType == valueTypePrimitiveArray {
				// primitive array
				if pathIdx == len(pathLvs)-1 {
					if idx < len(sub.primitiveArr) {
						return sub.primitiveArr[idx]
					} else {
						return nil
					}
				} else {
					//panic(errors.New("primitive array should be the last level"))
					return nil //FIXME should raise error?
				}
			} else {
				//panic(errors.New("not a array or primitive array"))
				return nil //FIXME should raise error?
			}
		}
	}
	if parent.valueType == valueTypePrimitive {
		return parent.value
	} else {
		return nil //FIXME should raise error?
	}
}

func (m *modelInst2MapImpl) RemoveObjectByPath(paths []string) error {
	if m.valueType != valueTypeObject {
		return errors.New("type is not object")
	}

	parent := m
	for _, v := range paths[:len(paths)-1] {
		lv, ok := parent.data[v]
		if !ok {
			// not found, stop
			return nil
		}
		if lv.valueType != valueTypeObject {
			return errors.New("sub type is not object")
		}
		parent = lv
	}
	delete(parent.data, paths[len(paths)-1])
	return nil
}

func (m *modelInst2MapImpl) putPrimitiveValue(name string, val interface{}) error {
	if m.valueType != valueTypeObject {
		return errors.New("type is not object")
	}
	if !isPrimitive(val) {
		return errors.New("value should be primitive")
	}
	m.data[name] = &modelInst2MapImpl{
		data:         nil,
		array:        nil,
		primitiveArr: nil,
		value:        val,
		valueType:    valueTypePrimitive,
	}
	return nil
}

func (m *modelInst2MapImpl) putPrimitiveArray(name string, arr []interface{}) error {
	if m.valueType != valueTypeObject {
		return errors.New("type is not object")
	}
	// check if value is primitive
	for _, v := range arr {
		if !isPrimitive(v) {
			return errors.New("value should be primitive")
		}
		break
	}
	var newArr = make([]interface{}, len(arr))
	copy(newArr, arr)
	// create sub primitive array
	sub := &modelInst2MapImpl{
		data:         nil,
		array:        nil,
		primitiveArr: newArr,
		value:        nil,
		valueType:    valueTypePrimitiveArray,
	}
	m.data[name] = sub
	return nil
}

func (m *modelInst2MapImpl) setPrimitiveArrayIndex(name string, index int, value interface{}) error {
	if m.valueType != valueTypeObject {
		return errors.New("type is not object")
	}
	if !isPrimitive(value) {
		return errors.New("value should be primitive")
	}
	subArr, ok := m.data[name]
	if !ok {
		defaultValue := defaultValuePrimitive(value)
		arr := make([]interface{}, index+1)
		for i, _ := range arr {
			arr[i] = defaultValue
		}
		arr[index] = value
		// create sub primitive array
		sub := &modelInst2MapImpl{
			data:         nil,
			array:        nil,
			primitiveArr: arr,
			value:        nil,
			valueType:    valueTypePrimitiveArray,
		}
		m.data[name] = sub
		return nil
	} else {
		if subArr.valueType != valueTypePrimitiveArray {
			return errors.New("sub type is not primitive array")
		}
		//FIXME need check value type matches elements in the array
		if index < len(subArr.primitiveArr) {
			subArr.primitiveArr[index] = value
			return nil
		} else {
			defaultValue := defaultValuePrimitive(value)
			for {
				subArr.primitiveArr = append(subArr.primitiveArr, defaultValue)
				if index < len(subArr.primitiveArr) {
					subArr.primitiveArr[index] = value
					return nil
				}
			}
		}
	}
}

func (m *modelInst2MapImpl) transferPrimitiveArray(srcName, dstName string, dst ModelInst2) error {
	if m.valueType != valueTypeObject {
		return errors.New("type is not object")
	}
	sub, ok := m.data[srcName]
	if !ok {
		return nil
	}
	if sub.valueType != valueTypePrimitiveArray {
		return errors.New(fmt.Sprintf("sub field=[%s] is not primitive array", srcName))
	}
	return dst.putPrimitiveArray(dstName, sub.primitiveArr)
}

func (m *modelInst2MapImpl) transferValue(srcName, dstName string, dst ModelInst2) error {
	if m.valueType != valueTypeObject {
		return errors.New("type is not object")
	}

	sub, ok := m.data[srcName]
	if !ok {
		return nil
	}
	if sub.valueType != valueTypePrimitive {
		return errors.New(fmt.Sprintf("sub field=[%s] is not primitive", srcName))
	}
	return dst.putPrimitiveValue(dstName, sub.value)
}

func (m *modelInst2MapImpl) ensureSubObject(name string) (ModelInst2, error) {
	if m.valueType != valueTypeObject {
		return nil, errors.New("type is not object")
	}

	sub, ok := m.data[name]
	if ok {
		if sub.valueType != valueTypeObject {
			return nil, errors.New(fmt.Sprintf("sub field=[%s] is not object", name))
		}
		return sub, nil
	}
	sub = &modelInst2MapImpl{
		data:         map[string]*modelInst2MapImpl{},
		array:        nil,
		primitiveArr: nil,
		value:        nil,
		valueType:    valueTypeObject,
	}
	m.data[name] = sub
	return sub, nil
}

func (m *modelInst2MapImpl) getSubObject(name string) (ModelInst2, error) {
	if m.valueType != valueTypeObject {
		return nil, errors.New("type is not object")
	}

	sub, ok := m.data[name]
	if ok {
		if sub.valueType != valueTypeObject {
			return nil, errors.New(fmt.Sprintf("sub field=[%s] is not object", name))
		}
		return sub, nil
	}
	return nil, nil
}

func (m *modelInst2MapImpl) ensureSubArrayWithObjectElem(name string) (ModelInst2, error) {
	if m.valueType != valueTypeObject {
		return nil, errors.New("type is not object")
	}

	sub, ok := m.data[name]
	if ok {
		if sub.valueType != valueTypeArray {
			return nil, errors.New(fmt.Sprintf("sub field=[%s] is not array", name))
		}
		return sub, nil
	}
	sub = &modelInst2MapImpl{
		data:         nil,
		array:        []*modelInst2MapImpl{},
		primitiveArr: nil,
		value:        nil,
		valueType:    valueTypeArray,
	}
	m.data[name] = sub
	return sub, nil
}

func (m *modelInst2MapImpl) getSubArrayWithObjectElem(name string) (ModelInst2, error) {
	if m.valueType != valueTypeObject {
		return nil, errors.New("type is not object")
	}

	sub, ok := m.data[name]
	if ok {
		if sub.valueType != valueTypeArray {
			return nil, errors.New(fmt.Sprintf("sub field=[%s] is not array", name))
		}
		return sub, nil
	}
	return nil, nil
}

func (m *modelInst2MapImpl) foreachArrayElement(f func(inst2 ModelInst2) error) error {
	if m.valueType != valueTypeArray {
		return errors.New("type is not array")
	}

	for _, v := range m.array {
		if err := f(v); err != nil {
			return err
		}
	}
	return nil
}

func (m *modelInst2MapImpl) ensureArrayElement() (ModelInst2, error) {
	if m.valueType != valueTypeArray {
		return nil, errors.New("type is not array")
	}

	newElem := &modelInst2MapImpl{
		data:         map[string]*modelInst2MapImpl{},
		array:        nil,
		primitiveArr: nil,
		value:        nil,
		valueType:    valueTypeObject,
	}
	m.array = append(m.array, newElem)
	return newElem, nil
}

func (m *modelInst2MapImpl) ensureArrayElementWithIndex(idx int) (ModelInst2, error) {
	if m.valueType != valueTypeArray {
		return nil, errors.New("type is not array")
	}

	if idx < len(m.array) {
		return m.array[idx], nil
	}

	for {
		newElem := &modelInst2MapImpl{
			data:         map[string]*modelInst2MapImpl{},
			array:        nil,
			primitiveArr: nil,
			value:        nil,
			valueType:    valueTypeObject,
		}
		m.array = append(m.array, newElem)
		if idx < len(m.array) {
			return newElem, nil
		}
	}
}
