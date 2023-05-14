package source

import (
	"github.com/ThisIsSun/fim/fimapi"
)

func InitSource(container fimapi.Container) error {
	if err := registerSourceConnectorGen(container, map[string]fimapi.SourceConnectorGenerator{
		"http_rest": sourceConnectorHttpRest,
	}); err != nil {
		return err
	}
	return nil
}

func registerSourceConnectorGen(container fimapi.Container, m map[string]fimapi.SourceConnectorGenerator) error {
	for name, connGen := range m {
		if err := container.RegisterSourceConnectorGen(name, connGen); err != nil {
			return err
		}
	}
	return nil
}
