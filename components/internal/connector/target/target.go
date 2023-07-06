package target

import (
	"github.com/FimGroup/fim/fimapi/pluginapi"
)

func InitTarget(a pluginapi.ApplicationSupport) error {
	if err := registerTargetConnectorGen(a, []pluginapi.TargetConnectorGenerator{
		NewDatabasePostgresGenerator(),
	}); err != nil {
		return err
	}
	return nil
}

func registerTargetConnectorGen(a pluginapi.ApplicationSupport, li []pluginapi.TargetConnectorGenerator) error {
	for _, connGen := range li {
		if err := a.AddTargetConnectorGenerator(connGen); err != nil {
			return err
		}
	}
	return nil
}
