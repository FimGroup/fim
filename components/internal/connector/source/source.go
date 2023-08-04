package source

import (
	"github.com/FimGroup/fim/components/internal/connector/source/scheduler"
	"github.com/FimGroup/fim/fimapi/pluginapi"
)

func InitSource(a pluginapi.ApplicationSupport) error {
	if err := registerSourceConnectorGen(a, []pluginapi.SourceConnectorGenerator{
		NewHttpRestServerGenerator(),
		scheduler.NewGoQuartzSchedulerSourceConnectorGenerator(),
	}); err != nil {
		return err
	}
	return nil
}

func registerSourceConnectorGen(a pluginapi.ApplicationSupport, li []pluginapi.SourceConnectorGenerator) error {
	for _, connGen := range li {
		if err := a.AddSourceConnectorGenerator(connGen); err != nil {
			return err
		}
	}
	return nil
}
