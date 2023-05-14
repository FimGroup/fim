package components

import (
	"errors"

	"github.com/ThisIsSun/fim/components/internal/connector/source"
	"github.com/ThisIsSun/fim/components/internal/connector/target"
	"github.com/ThisIsSun/fim/components/internal/fn"
	"github.com/ThisIsSun/fim/fimapi/basicapi"
	"github.com/ThisIsSun/fim/fimapi/pluginapi"
)

func InitComponent(c basicapi.BasicContainer) error {
	container, ok := c.(pluginapi.Container)
	if !ok {
		return errors.New("container type is not supported")
	}
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
