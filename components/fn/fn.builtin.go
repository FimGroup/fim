package fn

import (
	"errors"

	"esbconcept/esbcore"
)

func FnAssign(params []interface{}) (esbcore.Fn, error) {
	var field string = params[0].(string)
	if !esbcore.ValidateFullPath(field) {
		return nil, errors.New("path invalid:" + field)
	}
	fieldPaths := esbcore.SplitFullPath(field)
	var val = params[1]
	return func(m *esbcore.ModelInst) error {
		return m.AddOrUpdateField0(fieldPaths, val)
	}, nil
}
