package fimcore

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/pelletier/go-toml/v2"

	"github.com/ThisIsSun/fim/fimapi"
	"github.com/ThisIsSun/fim/fimapi/rule"
)

type templateFlowModel struct {
	Model map[string]string `toml:"model"`
}

var primitiveType = map[fimapi.DataType]struct{}{}

func init() {
	primitiveType[fimapi.DataTypeInt] = struct{}{}
	primitiveType[fimapi.DataTypeString] = struct{}{}
	primitiveType[fimapi.DataTypeBool] = struct{}{}
	primitiveType[fimapi.DataTypeFloat] = struct{}{}
}

type DataTypeDefinitions struct {
	fimapi.DataType
	PrimitiveArrayElementType fimapi.DataType // exists only when the element data type is primitive
	dataTypeMap               map[string]*DataTypeDefinitions
}

func NewDataTypeDefinitions() *DataTypeDefinitions {
	dtd := newInternalDataTypeDefinitions()
	dtd.DataType = fimapi.DataTypeObject
	return dtd
}

func newInternalDataTypeDefinitions() *DataTypeDefinitions {
	return &DataTypeDefinitions{
		dataTypeMap: map[string]*DataTypeDefinitions{},
	}
}

func newLastLevelDataTypeDefinitions() *DataTypeDefinitions {
	return &DataTypeDefinitions{}
}

func (d *DataTypeDefinitions) MergeToml(data string) error {
	m := new(templateFlowModel)
	err := toml.NewDecoder(bytes.NewBufferString(data)).DisallowUnknownFields().Decode(m)
	if err != nil {
		return err
	}

	return d.AddTypeDefinitions(m)
}

func (d *DataTypeDefinitions) AddTypeDefinitions(m *templateFlowModel) error {
	// validate
	for path, dataTypeStr := range m.Model {
		if !rule.ValidateFullPathOfDefinition(path) {
			return errors.New(fmt.Sprint("path:", path, " illegal"))
		}

		switch dataTypeStr {
		case "string":
		case "int":
		case "float":
		case "bool":
		default:
			return errors.New(fmt.Sprint("unknown dataType:", dataTypeStr))
		}
	}
	// process FlowModel items
	for k, v := range m.Model {
		if err := d.addTypeDefinitionOfPath(k, v); err != nil {
			return err
		}
	}
	return nil
}

func (d *DataTypeDefinitions) addTypeDefinitionOfPath(path string, dataTypeStr string) error {
	var dataType fimapi.DataType
	switch dataTypeStr {
	case "string":
		dataType = fimapi.DataTypeString
	case "int":
		dataType = fimapi.DataTypeInt
	case "float":
		dataType = fimapi.DataTypeFloat
	case "bool":
		dataType = fimapi.DataTypeBool
	default:
		return errors.New(fmt.Sprint("unknown dataType:", dataTypeStr))
	}

	paths := rule.SplitFullPath(path)
	objMap := d.dataTypeMap
	// process each level
	for _, pLv := range paths[:len(paths)-1] {
		isPathArr := rule.IsPathArray(pLv)
		if isPathArr {
			// extract path name
			name, _ := rule.ExtractArrayPath(pLv)
			pLv = name
		}
		lv, ok := objMap[pLv]
		if !ok {
			// not exist
			lv := newInternalDataTypeDefinitions()
			if isPathArr {
				lv.DataType = fimapi.DataTypeArray
			} else {
				lv.DataType = fimapi.DataTypeObject
			}
			objMap[pLv] = lv
			objMap = lv.dataTypeMap
			continue
		} else {
			// exist
			if isPathArr {
				if lv.DataType != fimapi.DataTypeArray {
					return errors.New(fmt.Sprintf("data type of path:%s is not array at level:%s", path, pLv))
				}
			} else {
				if lv.DataType != fimapi.DataTypeObject {
					return errors.New(fmt.Sprintf("data type of path:%s is not object at level:%s", path, pLv))
				}
			}
			objMap = lv.dataTypeMap
			continue
		}
	}
	// last level
	{
		lastLv := paths[len(paths)-1]
		isPathArr := rule.IsPathArray(lastLv)
		if isPathArr {
			// extract path name
			name, _ := rule.ExtractArrayPath(lastLv)
			lastLv = name
		}
		dtd, ok := objMap[lastLv]
		if !ok {
			// not exist
			dtd = newLastLevelDataTypeDefinitions()
			if isPathArr {
				dtd.DataType = fimapi.DataTypeArray
				dtd.PrimitiveArrayElementType = dataType
			} else {
				dtd.DataType = dataType
				dtd.PrimitiveArrayElementType = fimapi.DataTypeUnavailable
			}
			objMap[lastLv] = dtd
		} else {
			// exist
			return errors.New(fmt.Sprintf("duplicated definition of path:%s", path))
		}
	}

	return nil
}

// TypeOfPath returns the path data type, primitive array element data type and error
func (d *DataTypeDefinitions) TypeOfPath(path string) (fimapi.DataType, fimapi.DataType, error) {
	if !rule.ValidateFullPath(path) {
		return fimapi.DataTypeUnavailable, fimapi.DataTypeUnavailable, errors.New(fmt.Sprint("path:", path, " illegal"))
	}

	splits := rule.SplitFullPath(path)
	return d.typeOfPaths(splits)
}

func (d *DataTypeDefinitions) typeOfPaths(paths []string) (fimapi.DataType, fimapi.DataType, error) {
	dtd := d
	isAccessArrElem := false
	for _, pLv := range paths {
		lvName, arrIdx := rule.ExtractArrayPath(pLv) //FIXME need to identify primitive type or object/array type
		isAccessArrElem = arrIdx != -1
		pLv = lvName
		subDtd, ok := dtd.dataTypeMap[pLv]
		if !ok {
			return fimapi.DataTypeUnavailable, fimapi.DataTypeUnavailable, errors.New(fmt.Sprintf("path:%s not found", strings.Join(paths, fimapi.PathSeparator)))
		}
		dtd = subDtd
	}

	// handling last level if it is array related level
	dataType := dtd.DataType
	pDataType := dtd.PrimitiveArrayElementType
	if isAccessArrElem {
		_, isPrimitive := primitiveType[pDataType]
		if isPrimitive {
			dataType = pDataType
			pDataType = fimapi.DataTypeUnavailable
		} else {
			dataType = fimapi.DataTypeObject
			pDataType = fimapi.DataTypeUnavailable
		}
	}

	return dataType, pDataType, nil
}
