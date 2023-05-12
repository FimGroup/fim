package fn

import (
	"esbconcept/esbapi"
)

func InitFn(container esbapi.Container) error {

	if err := registerFn(container, map[string]esbapi.FnGen{
		"@assign": FnAssign,
		"@uuid":   FnUUID,
	}); err != nil {
		return err
	}

	return nil
}

func registerFn(container esbapi.Container, m map[string]esbapi.FnGen) error {
	for name, fn := range m {
		if err := container.RegisterBuiltinFn(name, fn); err != nil {
			return err
		}
	}
	return nil
}
