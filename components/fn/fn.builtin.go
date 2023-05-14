package fn

import (
	"errors"
	"time"

	"github.com/gofrs/uuid/v5"

	"github.com/ThisIsSun/fim/fimapi"
	"github.com/ThisIsSun/fim/fimapi/rule"
)

func FnAssign(params []interface{}) (fimapi.Fn, error) {
	var field string = params[0].(string)
	if !rule.ValidateFullPath(field) {
		return nil, errors.New("path invalid:" + field)
	}
	fieldPaths := rule.SplitFullPath(field)
	var val = params[1]
	return func(m fimapi.Model) error {
		return m.AddOrUpdateField0(fieldPaths, val)
	}, nil
}

func FnUUID(params []interface{}) (fimapi.Fn, error) {
	var field string = params[0].(string)
	if !rule.ValidateFullPath(field) {
		return nil, errors.New("path invalid:" + field)
	}
	fieldPaths := rule.SplitFullPath(field)
	return func(m fimapi.Model) error {
		u, err := uuid.NewV4()
		if err != nil {
			return err
		}
		return m.AddOrUpdateField0(fieldPaths, u.String())
	}, nil
}

func FnSetCurrentUnixTimestamp(params []interface{}) (fimapi.Fn, error) {
	var field string = params[0].(string)
	if !rule.ValidateFullPath(field) {
		return nil, errors.New("path invalid:" + field)
	}
	fieldPaths := rule.SplitFullPath(field)
	return func(m fimapi.Model) error {
		return m.AddOrUpdateField0(fieldPaths, int(time.Now().UnixMilli()))
	}, nil
}
