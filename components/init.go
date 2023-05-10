package components

import (
	"esbconcept/components/connector/source"
	"esbconcept/components/connector/target"
	"esbconcept/components/fn"
	"esbconcept/esbcore"
)

func InitComponent(container *esbcore.Container) error {
	if err := fn.InitFn(container); err != nil {
		return err
	}
	if err := source.InitSource(container); err != nil {
		return err
	}
	if err := target.InitTarget(container); err != nil {
		return err
	}
	return nil
}
