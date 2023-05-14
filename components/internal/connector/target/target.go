package target

import (
	"github.com/ThisIsSun/fim/fimapi/pluginapi"
)

func InitTarget(container pluginapi.Container) error {
	if err := registerTargetConnectorGen(container, []pluginapi.TargetConnectorGenerator{
		new(dbPgConnectorGenerator),
	}); err != nil {
		return err
	}
	return nil
}

func registerTargetConnectorGen(container pluginapi.Container, li []pluginapi.TargetConnectorGenerator) error {
	for _, connGen := range li {
		if err := container.RegisterTargetConnectorGen(connGen); err != nil {
			return err
		}
	}
	return nil
}
