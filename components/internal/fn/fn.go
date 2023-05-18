package fn

import (
	"github.com/ThisIsSun/fim/fimapi/pluginapi"
)

func InitFn(container pluginapi.Container) error {

	if err := registerFn(container, map[string]pluginapi.FnGen{
		"@assign":                     FnAssign,
		"@uuid":                       FnUUID,
		"@set_current_unix_timestamp": FnSetCurrentUnixTimestamp,

		"@crypto_bcrypt":        FnCryptoBcrypt,
		"@crypto_bcrypt_verify": FnCryptoBcryptVerify,

		"@check_always_break":    CheckAlwaysBreak,
		"@check_empty_break":     CheckEmptyBreak,
		"@check_not_blank_break": CheckNotBlankBreak,
	}); err != nil {
		return err
	}

	return nil
}

func registerFn(container pluginapi.Container, m map[string]pluginapi.FnGen) error {
	for name, fn := range m {
		if err := container.RegisterBuiltinFn(name, fn); err != nil {
			return err
		}
	}
	return nil
}
