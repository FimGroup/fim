package fn

import (
	"esbconcept/esbapi"
	"esbconcept/esbcore"
)

func InitFn(container *esbcore.Container) error {

	if err := registerFn(container, map[string]esbapi.FnGen{
		"@assign": FnAssign,
	}); err != nil {
		return err
	}

	return nil
}

func registerFn(container *esbcore.Container, m map[string]esbapi.FnGen) error {
	for name, fn := range m {
		if err := container.RegisterBuiltinFn(name, fn); err != nil {
			return err
		}
	}
	return nil
}
