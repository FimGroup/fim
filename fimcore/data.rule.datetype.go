package fimcore

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/pelletier/go-toml/v2"

	"github.com/FimGroup/fim/fimapi/pluginapi"
	"github.com/FimGroup/fim/fimapi/rule"
)

type templateFlowModel struct {
	Model map[string]string `toml:"model"`
}

var primitiveType = map[pluginapi.DataType]struct{}{}

func init() {
	primitiveType[pluginapi.DataTypeInt] = struct{}{}
	primitiveType[pluginapi.DataTypeString] = struct{}{}
	primitiveType[pluginapi.DataTypeBool] = struct{}{}
	primitiveType[pluginapi.DataTypeFloat] = struct{}{}
}

type DataTypeDefinitions struct {
	pluginapi.DataType
	PrimitiveArrayElementType pluginapi.DataType // exists only when the element data type is primitive
	dataTypeMap               map[string]*DataTypeDefinitions
}

func NewDataTypeDefinitions() *DataTypeDefinitions {
	dtd := newInternalDataTypeDefinitions()
	dtd.DataType = pluginapi.DataTypeObject
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
	var dataType pluginapi.DataType
	switch dataTypeStr {
	case "string":
		dataType = pluginapi.DataTypeString
	case "int":
		dataType = pluginapi.DataTypeInt
	case "float":
		dataType = pluginapi.DataTypeFloat
	case "bool":
		dataType = pluginapi.DataTypeBool
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
				lv.DataType = pluginapi.DataTypeArray
			} else {
				lv.DataType = pluginapi.DataTypeObject
			}
			objMap[pLv] = lv
			objMap = lv.dataTypeMap
			continue
		} else {
			// exist
			if isPathArr {
				if lv.DataType != pluginapi.DataTypeArray {
					return errors.New(fmt.Sprintf("data type of path:%s is not array at level:%s", path, pLv))
				}
			} else {
				if lv.DataType != pluginapi.DataTypeObject {
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
				dtd.DataType = pluginapi.DataTypeArray
				dtd.PrimitiveArrayElementType = dataType
			} else {
				dtd.DataType = dataType
				dtd.PrimitiveArrayElementType = pluginapi.DataTypeUnavailable
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
func (d *DataTypeDefinitions) TypeOfPath(path string) (pluginapi.DataType, pluginapi.DataType, error) {
	if !rule.ValidateFullPath(path) {
		return pluginapi.DataTypeUnavailable, pluginapi.DataTypeUnavailable, errors.New(fmt.Sprint("path:", path, " illegal"))
	}

	splits := rule.SplitFullPath(path)
	return d.typeOfPaths(splits)
}

func (d *DataTypeDefinitions) typeOfPaths(paths []string) (pluginapi.DataType, pluginapi.DataType, error) {
	dtd := d
	isAccessArrElem := false
	for _, pLv := range paths {
		lvName, arrIdx := rule.ExtractArrayPath(pLv) //FIXME need to identify primitive type or object/array type
		isAccessArrElem = arrIdx != -1
		pLv = lvName
		subDtd, ok := dtd.dataTypeMap[pLv]
		if !ok {
			return pluginapi.DataTypeUnavailable, pluginapi.DataTypeUnavailable, errors.New(fmt.Sprintf("path:%s not found", strings.Join(paths, pluginapi.PathSeparator)))
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
			pDataType = pluginapi.DataTypeUnavailable
		} else {
			dataType = pluginapi.DataTypeObject
			pDataType = pluginapi.DataTypeUnavailable
		}
	}

	return dataType, pDataType, nil
}
