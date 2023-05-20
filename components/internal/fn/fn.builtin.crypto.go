package fn

import (
	"errors"

	"golang.org/x/crypto/bcrypt"

	"github.com/FimGroup/fim/fimapi/basicapi"
	"github.com/FimGroup/fim/fimapi/pluginapi"
	"github.com/FimGroup/fim/fimapi/rule"
)

func FnCryptoBcrypt(params []interface{}) (pluginapi.Fn, error) {
	var field string = params[0].(string)
	if !rule.ValidateFullPath(field) {
		return nil, errors.New("path invalid:" + field)
	}
	fieldPaths := rule.SplitFullPath(field)
	return func(m pluginapi.Model) error {
		val := m.GetFieldUnsafe(fieldPaths)
		if val == nil {
			return nil
		}
		sval, ok := val.(string)
		if !ok {
			return errors.New("data type is not string")
		}
		data, err := bcrypt.GenerateFromPassword([]byte(sval), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		return m.AddOrUpdateField0(fieldPaths, string(data))
	}, nil
}

func FnCryptoBcryptVerify(params []interface{}) (pluginapi.Fn, error) {
	var field string = params[0].(string)
	if !rule.ValidateFullPath(field) {
		return nil, errors.New("path invalid:" + field)
	}
	bcryptoDataPaths := rule.SplitFullPath(field)
	var userInputField string = params[1].(string)
	if !rule.ValidateFullPath(userInputField) {
		return nil, errors.New("path invalid:" + userInputField)
	}
	userInputDataPaths := rule.SplitFullPath(userInputField)
	var validateResultPath string = params[2].(string)
	if !rule.ValidateFullPath(validateResultPath) {
		return nil, errors.New("path invalid:" + validateResultPath)
	}
	validateResultPaths := rule.SplitFullPath(validateResultPath)
	return func(m basicapi.Model) error {
		val := m.GetFieldUnsafe(bcryptoDataPaths)
		if val == nil {
			return nil
		}
		sval, ok := val.(string)
		if !ok {
			return errors.New("data type is not string")
		}
		inputVal := m.GetFieldUnsafe(userInputDataPaths)
		if inputVal == nil {
			return nil
		}
		sinputVal, ok := inputVal.(string)
		if !ok {
			return errors.New("data type is not string")
		}
		err := bcrypt.CompareHashAndPassword([]byte(sval), []byte(sinputVal))
		if err == nil {
			return m.AddOrUpdateField0(validateResultPaths, true)
		} else if err == bcrypt.ErrMismatchedHashAndPassword {
			return m.AddOrUpdateField0(validateResultPaths, false)
		} else {
			return err
		}
	}, nil
}
