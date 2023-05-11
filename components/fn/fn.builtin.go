package fn

import (
	"errors"

	"esbconcept/esbapi"
	"esbconcept/esbcore"
)

func FnAssign(params []interface{}) (esbapi.Fn, error) {
	var field string = params[0].(string)
	if !esbcore.ValidateFullPath(field) {
		return nil, errors.New("path invalid:" + field)
	}
	fieldPaths := esbcore.SplitFullPath(field)
	var val = params[1]
	return func(m esbapi.Model) error {
		return m.(*esbcore.ModelInst).AddOrUpdateField0(fieldPaths, val)
	}, nil
}
