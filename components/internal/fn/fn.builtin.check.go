package fn

import (
	"errors"
	"strings"

	"github.com/FimGroup/fim/fimapi/pluginapi"
	"github.com/FimGroup/fim/fimapi/rule"
)

func CheckAlwaysBreak(params []interface{}) (pluginapi.Fn, error) {
	errorKey := params[0].(string)
	errorMessage := params[1].(string)
	return func(m pluginapi.Model) error {
		return &pluginapi.FlowError{
			Key:     errorKey,
			Message: errorMessage,
		}
	}, nil
}

func CheckNotExistBreak(params []interface{}) (pluginapi.Fn, error) {
	var field string = params[0].(string)
	if !rule.ValidateFullPath(field) {
		return nil, errors.New("path invalid:" + field)
	}
	fieldPaths := rule.SplitFullPath(field)
	errorKey := params[1].(string)
	errorMessage := params[2].(string)
	return func(m pluginapi.Model) error {
		val := m.GetFieldUnsafe0(fieldPaths)
		if val == nil {
			return nil
		}
		return &pluginapi.FlowError{
			Key:     errorKey,
			Message: errorMessage,
		}
	}, nil
}

func CheckExistBreak(params []interface{}) (pluginapi.Fn, error) {
	var field string = params[0].(string)
	if !rule.ValidateFullPath(field) {
		return nil, errors.New("path invalid:" + field)
	}
	fieldPaths := rule.SplitFullPath(field)
	errorKey := params[1].(string)
	errorMessage := params[2].(string)
	return func(m pluginapi.Model) error {
		val := m.GetFieldUnsafe0(fieldPaths)
		if val != nil {
			return nil
		}
		return &pluginapi.FlowError{
			Key:     errorKey,
			Message: errorMessage,
		}
	}, nil
}

func CheckEmptyBreak(params []interface{}) (pluginapi.Fn, error) {
	var field string = params[0].(string)
	if !rule.ValidateFullPath(field) {
		return nil, errors.New("path invalid:" + field)
	}
	fieldPaths := rule.SplitFullPath(field)
	errorKey := params[1].(string)
	errorMessage := params[2].(string)
	return func(m pluginapi.Model) error {
		val := m.GetFieldUnsafe0(fieldPaths)
		if val == nil {
			return nil
		}
		s, ok := val.(string)
		if !ok {
			return errors.New("CheckEmptyBreak: data type is not string")
		}
		if s == "" {
			return nil
		}
		return &pluginapi.FlowError{
			Key:     errorKey,
			Message: errorMessage,
		}
	}, nil
}

func CheckNotBlankBreak(params []interface{}) (pluginapi.Fn, error) {
	var field string = params[0].(string)
	if !rule.ValidateFullPath(field) {
		return nil, errors.New("path invalid:" + field)
	}
	fieldPaths := rule.SplitFullPath(field)
	errorKey := params[1].(string)
	errorMessage := params[2].(string)
	return func(m pluginapi.Model) error {
		val := m.GetFieldUnsafe0(fieldPaths)
		if val == nil {
			return &pluginapi.FlowError{
				Key:     errorKey,
				Message: errorMessage,
			}
		}
		s, ok := val.(string)
		if !ok {
			return errors.New("CheckNotBlankBreak: data type is not string")
		}
		if strings.TrimSpace(s) != "" {
			return nil
		}
		return &pluginapi.FlowError{
			Key:     errorKey,
			Message: errorMessage,
		}
	}, nil
}
