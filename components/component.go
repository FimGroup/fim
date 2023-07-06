package components

import (
	"errors"

	"github.com/FimGroup/fim/components/internal/connector/source"
	"github.com/FimGroup/fim/components/internal/connector/target"
	"github.com/FimGroup/fim/components/internal/fn"
	"github.com/FimGroup/fim/fimapi/basicapi"
	"github.com/FimGroup/fim/fimapi/pluginapi"
)

func InitConnectors(a pluginapi.ApplicationSupport) error {
	if err := source.InitSource(a); err != nil {
		return err
	}
	if err := target.InitTarget(a); err != nil {
		return err
	}
	return nil
}

func InitFunctions(c basicapi.BasicContainer) error {
	container, ok := c.(pluginapi.Container)
	if !ok {
		return errors.New("container type is not supported")
	}
	if err := fn.InitFn(container); err != nil {
		return err
	}
	return nil
}
