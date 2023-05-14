package components

import (
	"github.com/ThisIsSun/fim/components/connector/source"
	"github.com/ThisIsSun/fim/components/connector/target"
	"github.com/ThisIsSun/fim/components/fn"
	"github.com/ThisIsSun/fim/fimapi"
)

func InitComponent(container fimapi.Container) error {
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
