package fn

import (
	"github.com/ThisIsSun/fim/fimapi"
)

func InitFn(container fimapi.Container) error {

	if err := registerFn(container, map[string]fimapi.FnGen{
		"@assign":                     FnAssign,
		"@uuid":                       FnUUID,
		"@set_current_unix_timestamp": FnSetCurrentUnixTimestamp,
	}); err != nil {
		return err
	}

	return nil
}

func registerFn(container fimapi.Container, m map[string]fimapi.FnGen) error {
	for name, fn := range m {
		if err := container.RegisterBuiltinFn(name, fn); err != nil {
			return err
		}
	}
	return nil
}
