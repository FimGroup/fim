package esbcore

import (
	"errors"

	"esbconcept/esbapi"
)

var customGenFnMap map[string]func(params []interface{}) (esbapi.Fn, error)

func init() {
	customGenFnMap = map[string]func(params []interface{}) (esbapi.Fn, error){}
}

func RegisterCustomGeneratorFunc(name string, fn func(params []interface{}) (esbapi.Fn, error)) error {
	_, ok := customGenFnMap[name]
	if ok {
		return errors.New("custom function already exists:" + name)
	}
	customGenFnMap[name] = fn
	return nil
}
