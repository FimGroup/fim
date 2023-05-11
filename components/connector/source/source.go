package source

import (
	"esbconcept/esbapi"
)

func InitSource(container esbapi.Container) error {
	if err := registerSourceConnectorGen(container, map[string]esbapi.SourceConnectorGenerator{
		"http_rest": sourceConnectorHttpRest,
	}); err != nil {
		return err
	}
	return nil
}

func registerSourceConnectorGen(container esbapi.Container, m map[string]esbapi.SourceConnectorGenerator) error {
	for name, connGen := range m {
		if err := container.RegisterSourceConnectorGen(name, connGen); err != nil {
			return err
		}
	}
	return nil
}
