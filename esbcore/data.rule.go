package esbcore

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/pelletier/go-toml/v2"
)

type templateFlowModel struct {
	Model map[string]string `toml:"model"`
}

type DataType int

const (
	DataTypeUnavailable DataType = 0
	DataTypeInt         DataType = 1
	DataTypeString      DataType = 2
	DataTypeBool        DataType = 3
	DataTypeFloat       DataType = 4
	DataTypeArray       DataType = 11
	DataTypeObject      DataType = 12
)

type DataTypeDefinitions struct {
	DataType
	PrimitiveArrayElementType DataType // exists only when the element data type is primitive
	dataTypeMap               map[string]*DataTypeDefinitions
}

func NewDataTypeDefinitions() *DataTypeDefinitions {
	dtd := newInternalDataTypeDefinitions()
	dtd.DataType = DataTypeObject
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
		if !ValidateFullPathOfDefinition(path) {
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
	var dataType DataType
	switch dataTypeStr {
	case "string":
		dataType = DataTypeString
	case "int":
		dataType = DataTypeInt
	case "float":
		dataType = DataTypeFloat
	case "bool":
		dataType = DataTypeBool
	default:
		return errors.New(fmt.Sprint("unknown dataType:", dataTypeStr))
	}

	paths := SplitFullPath(path)
	objMap := d.dataTypeMap
	// process each level
	for _, pLv := range paths[:len(paths)-1] {
		isPathArr := IsPathArray(pLv)
		if isPathArr {
			// extract path name
			name, _ := ExtractArrayPath(pLv)
			pLv = name
		}
		lv, ok := objMap[pLv]
		if !ok {
			// not exist
			lv := newInternalDataTypeDefinitions()
			if isPathArr {
				lv.DataType = DataTypeArray
			} else {
				lv.DataType = DataTypeObject
			}
			objMap[pLv] = lv
			objMap = lv.dataTypeMap
			continue
		} else {
			// exist
			if isPathArr {
				if lv.DataType != DataTypeArray {
					return errors.New(fmt.Sprintf("data type of path:%s is not array at level:%s", path, pLv))
				}
			} else {
				if lv.DataType != DataTypeObject {
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
		isPathArr := IsPathArray(lastLv)
		if isPathArr {
			// extract path name
			name, _ := ExtractArrayPath(lastLv)
			lastLv = name
		}
		dtd, ok := objMap[lastLv]
		if !ok {
			// not exist
			dtd = newLastLevelDataTypeDefinitions()
			if isPathArr {
				dtd.DataType = DataTypeArray
				dtd.PrimitiveArrayElementType = dataType
			} else {
				dtd.DataType = dataType
				dtd.PrimitiveArrayElementType = DataTypeUnavailable
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
func (d *DataTypeDefinitions) TypeOfPath(path string) (DataType, DataType, error) {
	if !ValidateFullPath(path) {
		return DataTypeUnavailable, DataTypeUnavailable, errors.New(fmt.Sprint("path:", path, " illegal"))
	}

	splits := SplitFullPath(path)
	dtd := d
	isAccessArrElem := false
	for _, pLv := range splits {
		lvName, arrIdx := ExtractArrayPath(pLv) //FIXME need to identify primitive type or object/array type
		isAccessArrElem = arrIdx != -1
		pLv = lvName
		subDtd, ok := dtd.dataTypeMap[pLv]
		if !ok {
			return DataTypeUnavailable, DataTypeUnavailable, errors.New(fmt.Sprintf("path:%s not found", path))
		}
		dtd = subDtd
	}

	dataType := dtd.DataType
	pDataType := dtd.PrimitiveArrayElementType
	if isAccessArrElem {
		dataType = pDataType
		pDataType = DataTypeUnavailable
	}

	return dataType, pDataType, nil
}
