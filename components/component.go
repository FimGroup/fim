package components

import (
	"github.com/ThisIsSun/fim/components/connector/source"
	"github.com/ThisIsSun/fim/components/connector/target"
	"github.com/ThisIsSun/fim/components/fn"
	"github.com/ThisIsSun/fim/fimapi/pluginapi"
)

func InitComponent(container pluginapi.Container) error {
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
