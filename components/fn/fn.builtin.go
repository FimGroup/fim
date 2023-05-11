package fn

import (
	"errors"

	"esbconcept/esbapi"
	"esbconcept/esbapi/rule"
)

func FnAssign(params []interface{}) (esbapi.Fn, error) {
	var field string = params[0].(string)
	if !rule.ValidateFullPath(field) {
		return nil, errors.New("path invalid:" + field)
	}
	fieldPaths := rule.SplitFullPath(field)
	var val = params[1]
	return func(m esbapi.Model) error {
		return m.AddOrUpdateField0(fieldPaths, val)
	}, nil
}
