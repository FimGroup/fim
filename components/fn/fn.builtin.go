package fn

import (
	"errors"

	"github.com/gofrs/uuid/v5"

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

func FnUUID(params []interface{}) (esbapi.Fn, error) {
	var field string = params[0].(string)
	if !rule.ValidateFullPath(field) {
		return nil, errors.New("path invalid:" + field)
	}
	fieldPaths := rule.SplitFullPath(field)
	return func(m esbapi.Model) error {
		u, err := uuid.NewV4()
		if err != nil {
			return err
		}
		return m.AddOrUpdateField0(fieldPaths, u.String())
	}, nil
}
