package target

import (
	"github.com/ThisIsSun/fim/fimapi/pluginapi"
)

func InitTarget(container pluginapi.Container) error {
	if err := registerTargetConnectorGen(container, map[string]pluginapi.TargetConnectorGenerator{
		"&database_postgres": databasePostgresGenerator,
	}); err != nil {
		return err
	}
	return nil
}

func registerTargetConnectorGen(container pluginapi.Container, m map[string]pluginapi.TargetConnectorGenerator) error {
	for name, connGen := range m {
		if err := container.RegisterTargetConnectorGen(name, connGen); err != nil {
			return err
		}
	}
	return nil
}
