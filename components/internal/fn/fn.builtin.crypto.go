package fn

import (
	"errors"

	"golang.org/x/crypto/bcrypt"

	"github.com/ThisIsSun/fim/fimapi/pluginapi"
	"github.com/ThisIsSun/fim/fimapi/rule"
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
