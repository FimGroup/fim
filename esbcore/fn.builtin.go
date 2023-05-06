package esbcore

import "errors"

type Fn func(m *ModelInst) error

var builtinGenFnMap map[string]func(params []interface{}) (Fn, error)

func init() {
	builtinGenFnMap = map[string]func(params []interface{}) (Fn, error){}
	builtinGenFnMap["@assign"] = FnAssign
}

func FnAssign(params []interface{}) (Fn, error) {
	var field string = params[0].(string)
	if !ValidateFullPath(field) {
		return nil, errors.New("path invalid:" + field)
	}
	fieldPaths := SplitFullPath(field)
	var val = params[1]
	return func(m *ModelInst) error {
		return m.addOrUpdateField(fieldPaths, val)
	}, nil
}
