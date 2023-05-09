package esbcore

import "errors"

var customGenFnMap map[string]func(params []interface{}) (Fn, error)

func init() {
	customGenFnMap = map[string]func(params []interface{}) (Fn, error){}
}

func RegisterCustomGeneratorFunc(name string, fn func(params []interface{}) (Fn, error)) error {
	_, ok := customGenFnMap[name]
	if ok {
		return errors.New("custom function already exists:" + name)
	}
	customGenFnMap[name] = fn
	return nil
}
