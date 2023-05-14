package source

import (
	"github.com/ThisIsSun/fim/fimapi/pluginapi"
)

func InitSource(container pluginapi.Container) error {
	if err := registerSourceConnectorGen(container, []pluginapi.SourceConnectorGenerator{
		httpRestServer,
	}); err != nil {
		return err
	}
	return nil
}

func registerSourceConnectorGen(container pluginapi.Container, li []pluginapi.SourceConnectorGenerator) error {
	for _, connGen := range li {
		if err := container.RegisterSourceConnectorGen(connGen); err != nil {
			return err
		}
	}
	return nil
}
