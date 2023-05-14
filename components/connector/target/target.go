package target

import (
	"github.com/ThisIsSun/fim/fimapi"
)

func InitTarget(container fimapi.Container) error {
	if err := registerTargetConnectorGen(container, map[string]fimapi.TargetConnectorGenerator{
		"&database_postgres": databasePostgresGenerator,
	}); err != nil {
		return err
	}
	return nil
}

func registerTargetConnectorGen(container fimapi.Container, m map[string]fimapi.TargetConnectorGenerator) error {
	for name, connGen := range m {
		if err := container.RegisterTargetConnectorGen(name, connGen); err != nil {
			return err
		}
	}
	return nil
}
