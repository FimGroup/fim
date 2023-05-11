package esbprovider

import (
	"errors"
	"fmt"

	"esbconcept/esbapi"
	"esbconcept/esbcore"
)

func FnPrintObject(params []interface{}) (esbapi.Fn, error) {
	key := params[0].(string)
	if !esbcore.ValidateFullPath(key) {
		return nil, errors.New("")
	}
	paths := esbcore.SplitFullPath(key)
	return func(m esbapi.Model) error {
		//FIXME have to handle object/array properly
		o := m.(*esbcore.ModelInst).GetFieldUnsafe(paths)
		fmt.Println("print object:", o)
		return nil
	}, nil
}
