package fn

import "esbconcept/esbcore"

func InitFn(container *esbcore.Container) error {

	if err := registerFn(container, map[string]func(params []interface{}) (esbcore.Fn, error){
		"@assign": FnAssign,
	}); err != nil {
		return err
	}

	return nil
}

func registerFn(container *esbcore.Container, m map[string]func(params []interface{}) (esbcore.Fn, error)) error {
	for name, fn := range m {
		if err := container.RegisterBuiltinFn(name, fn); err != nil {
			return err
		}
	}
	return nil
}
