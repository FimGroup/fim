package sample

import (
	"errors"
	"fmt"

	"github.com/ThisIsSun/fim/fimapi"
	"github.com/ThisIsSun/fim/fimapi/rule"
)

func FnPrintObject(params []interface{}) (fimapi.Fn, error) {
	key := params[0].(string)
	if !rule.ValidateFullPath(key) {
		return nil, errors.New("invalid path:" + key)
	}
	paths := rule.SplitFullPath(key)
	return func(m fimapi.Model) error {
		//FIXME have to handle object/array properly
		o := m.GetFieldUnsafe(paths)
		fmt.Println("print object:", o)
		return nil
	}, nil
}
