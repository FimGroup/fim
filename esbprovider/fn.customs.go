package esbprovider

import (
	"errors"
	"fmt"

	"esbconcept/esbapi"
	"esbconcept/esbapi/rule"
)

func FnPrintObject(params []interface{}) (esbapi.Fn, error) {
	key := params[0].(string)
	if !rule.ValidateFullPath(key) {
		return nil, errors.New("")
	}
	paths := rule.SplitFullPath(key)
	return func(m esbapi.Model) error {
		//FIXME have to handle object/array properly
		o := m.GetFieldUnsafe(paths)
		fmt.Println("print object:", o)
		return nil
	}, nil
}
