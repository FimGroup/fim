package source

import (
	"github.com/ThisIsSun/fim/fimapi/pluginapi"
)

func InitSource(container pluginapi.Container) error {
	if err := registerSourceConnectorGen(container, map[string]pluginapi.SourceConnectorGenerator{
		"http_rest": sourceConnectorHttpRest,
	}); err != nil {
		return err
	}
	return nil
}

func registerSourceConnectorGen(container pluginapi.Container, m map[string]pluginapi.SourceConnectorGenerator) error {
	for name, connGen := range m {
		if err := container.RegisterSourceConnectorGen(name, connGen); err != nil {
			return err
		}
	}
	return nil
}
