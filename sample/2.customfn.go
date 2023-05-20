package sample

import (
	"errors"
	"fmt"

	"github.com/FimGroup/fim/fimapi/pluginapi"
	"github.com/FimGroup/fim/fimapi/rule"
)

func FnPrintObject(params []interface{}) (pluginapi.Fn, error) {
	key := params[0].(string)
	if !rule.ValidateFullPath(key) {
		return nil, errors.New("invalid path:" + key)
	}
	paths := rule.SplitFullPath(key)
	return func(m pluginapi.Model) error {
		//FIXME have to handle object/array properly
		o := m.GetFieldUnsafe(paths)
		fmt.Println("print object:", o)
		return nil
	}, nil
}

func FnPanic(params []interface{}) (pluginapi.Fn, error) {
	key := params[0].(string)
	return func(m pluginapi.Model) error {
		panic(errors.New(key))
	}, nil
}
