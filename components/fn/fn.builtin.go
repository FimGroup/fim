package fn

import (
	"errors"
	"time"

	"github.com/gofrs/uuid/v5"

	"github.com/ThisIsSun/fim/fimapi/pluginapi"
	"github.com/ThisIsSun/fim/fimapi/pluginapi/rule"
)

func FnAssign(params []interface{}) (pluginapi.Fn, error) {
	var field string = params[0].(string)
	if !rule.ValidateFullPath(field) {
		return nil, errors.New("path invalid:" + field)
	}
	fieldPaths := rule.SplitFullPath(field)
	var val = params[1]
	return func(m pluginapi.Model) error {
		return m.AddOrUpdateField0(fieldPaths, val)
	}, nil
}

func FnUUID(params []interface{}) (pluginapi.Fn, error) {
	var field string = params[0].(string)
	if !rule.ValidateFullPath(field) {
		return nil, errors.New("path invalid:" + field)
	}
	fieldPaths := rule.SplitFullPath(field)
	return func(m pluginapi.Model) error {
		u, err := uuid.NewV4()
		if err != nil {
			return err
		}
		return m.AddOrUpdateField0(fieldPaths, u.String())
	}, nil
}

func FnSetCurrentUnixTimestamp(params []interface{}) (pluginapi.Fn, error) {
	var field string = params[0].(string)
	if !rule.ValidateFullPath(field) {
		return nil, errors.New("path invalid:" + field)
	}
	fieldPaths := rule.SplitFullPath(field)
	return func(m pluginapi.Model) error {
		return m.AddOrUpdateField0(fieldPaths, int(time.Now().UnixMilli()))
	}, nil
}
