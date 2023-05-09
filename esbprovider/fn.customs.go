package esbprovider

import (
	"errors"
	"esbconcept/esbcore"
	"fmt"
)

func FnPrintObject(params []interface{}) (esbcore.Fn, error) {
	key := params[0].(string)
	if !esbcore.ValidateFullPath(key) {
		return nil, errors.New("")
	}
	paths := esbcore.SplitFullPath(key)
	return func(m *esbcore.ModelInst) error {
		//FIXME have to handle object/array properly
		o := m.GetFieldUnsafe(paths)
		fmt.Println("print object:", o)
		return nil
	}, nil
}
