package target

import (
	"esbconcept/esbapi"
)

func InitTarget(container esbapi.Container) error {
	if err := registerTargetConnectorGen(container, map[string]esbapi.TargetConnectorGenerator{
		"&database_postgres": databasePostgresGenerator,
	}); err != nil {
		return err
	}
	return nil
}

func registerTargetConnectorGen(container esbapi.Container, m map[string]esbapi.TargetConnectorGenerator) error {
	for name, connGen := range m {
		if err := container.RegisterTargetConnectorGen(name, connGen); err != nil {
			return err
		}
	}
	return nil
}
