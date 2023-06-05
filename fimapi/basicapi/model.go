package basicapi

import (
	"errors"
	"fmt"
	"reflect"
)

type Model interface {
	//FIXME need design the type and behaviors of Model

	//AddOrUpdateField0 supports primitive value only from object and array(nested array is ok)
	AddOrUpdateField0(path []string, value interface{}) error
	//GetFieldUnsafe0 supports primitive value only from object and array(nested array is ok)
	GetFieldUnsafe0(path []string) interface{}

	ToGeneralObject() interface{}
}

func ConvertPrimitive(in interface{}) (interface{}, error) {
	switch v := in.(type) {
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	case bool:
		return v, nil
	case uint8:
		return int64(v), nil
	case uint16:
		return int64(v), nil
	case uint32:
		return int64(v), nil
	case uint64:
		return nil, errors.New("currently uint64 is not supported for conversion")
	case int8:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	case complex64:
		return nil, errors.New("currently complex64 is not supported for conversion")
	case complex128:
		return nil, errors.New("currently complex128 is not supported for conversion")
	case string:
		return v, nil
	case int:
		return int64(v), nil
	case uint:
		return nil, errors.New("currently uint is not supported for conversion")
	case uintptr:
		return nil, errors.New("currently uintptr is not supported for conversion")
	}
	return nil, errors.New("unknown type for conversion:" + fmt.Sprint(reflect.TypeOf(in)))
}

func MustConvertPrimitive(in interface{}) interface{} {
	v, err := ConvertPrimitive(in)
	if err != nil {
		panic(err)
	}
	return v
}
