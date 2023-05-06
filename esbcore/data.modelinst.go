package esbcore

import (
	"errors"
	"fmt"
)

type TypeOfNode int
type ElementMap map[string]interface{}
type refBy int

const (
	ByLeft  refBy = 1
	ByRight refBy = 2
)

const (
	TypeUnknown       TypeOfNode = 0
	TypeDataNode      TypeOfNode = 1
	TypeNsNode        TypeOfNode = 2
	TypeAttributeNode TypeOfNode = 3
)

type ModelInst struct {
	dtd        *DataTypeDefinitions
	ElementMap ElementMap
}

func NewModelInst(def *DataTypeDefinitions) *ModelInst {
	return &ModelInst{
		dtd:        def,
		ElementMap: ElementMap{},
	}
}

func (m *ModelInst) addOrUpdateField(path string, value interface{}) error {
	splits := SplitFullPath(path)

	var result ElementMap = m.ElementMap
	for _, pLv := range splits[:len(splits)-1] { // 0 to second last level
		pathName, idx := ExtractArrayPath(pLv)
		isArrAccess := idx >= 0
		elemMap := result
		elem, ok := elemMap[pathName]
		if ok {
			result = elem.(ElementMap)
		} else {
			elem = ElementMap{}
			elemMap[pathName] = elem
			result = elem.(ElementMap)
		}
		// additional: process arr access
		if isArrAccess {
			elem, ok := result[fmt.Sprint(idx)]
			if !ok {
				elem = ElementMap{}
				result[fmt.Sprint(idx)] = elem
			}
			result = elem.(ElementMap)
		}
	}

	// last level
	{
		pathName, idx := ExtractArrayPath(splits[len(splits)-1])
		isArrAccess := idx >= 0
		if !isArrAccess {
			result[pathName] = value
		} else {
			arr, ok := result[pathName]
			if !ok {
				arr = ElementMap{}
				result[pathName] = arr
			}
			arr.(ElementMap)[fmt.Sprint(idx)] = value
		}
	}

	return nil
}

func (m *ModelInst) deleteField(path string) error {
	panic(_IMPLEMENT_ME)
}

func (m *ModelInst) getField(path string) interface{} {
	splits := SplitFullPath(path)

	var result interface{} = m.ElementMap
	for _, pLv := range splits {
		pathName, idx := ExtractArrayPath(pLv)
		isArrAccess := idx >= 0
		elemMap, ok := result.(ElementMap)
		if !ok {
			return nil
		}
		elem, ok := elemMap[pathName]
		if !ok {
			return nil
		}
		if isArrAccess {
			result = elem.(ElementMap)[fmt.Sprint(idx)]
		} else {
			result = elem
		}
	}

	return result
}

func (m *ModelInst) FillInFrom(o interface{}) error {
	panic(_IMPLEMENT_ME)
}

func (m *ModelInst) ExtractTo(o interface{}) error {
	panic(_IMPLEMENT_ME)
}

func (m *ModelInst) transferTo(dest *ModelInst, sourcePath, destPath string, defaultTypeRefBy refBy) error {
	val := m.getField(sourcePath)
	if val != nil {
		return dest.addOrUpdateField(destPath, val)
	} else {
		// default value handling when not existing
		var d DataType
		switch defaultTypeRefBy {
		case ByLeft:
			dt, _, err := m.dtd.TypeOfPath(sourcePath)
			if err != nil {
				return err
			}
			d = dt
		case ByRight:
			dt, _, err := m.dtd.TypeOfPath(destPath)
			if err != nil {
				return err
			}
			d = dt
		default:
			return errors.New("unknown refBy:" + fmt.Sprint(defaultTypeRefBy))
		}
		switch d {
		case DataTypeInt:
			return dest.addOrUpdateField(destPath, 0)
		case DataTypeString:
			return dest.addOrUpdateField(destPath, "")
		case DataTypeFloat:
			return dest.addOrUpdateField(destPath, 0.0)
		case DataTypeBool:
			return dest.addOrUpdateField(destPath, false)
		case DataTypeObject:
			return dest.deleteField(destPath)
		case DataTypeArray:
			return dest.deleteField(destPath)
		}
	}
	//FIXME use deep copy instead of setting reference to avoid modification issue
	//FIXME verify field type according to FlowModel
	return nil
}
