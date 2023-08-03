package fimcore

import (
	"errors"
	"fmt"
)

func (c *ContainerInst) LoadMerged(content string) error {
	m, err := LoadMergedDefinition(content)
	if err != nil {
		return err
	}
	return c.loadMerged0(m)
}

func (c *ContainerInst) loadMerged0(m *MergedDefinition) error {
	// load flow
	for name, tf := range m.Flows {
		_, ok := c.flowMap[name]
		if ok {
			return errors.New(fmt.Sprint("flow exists:", name))
		}

		flow := NewFlow(c.flowModel, c)
		if err := flow.mergeToml(tf); err != nil {
			return err
		}
		c.flowMap[name] = flow
		c.flowRawMap[name] = struct{ tf *templateFlow }{tf: tf}
	}

	// load pipeline
	for name, pipeline := range m.Pipelines {
		_, ok := c.pipelineMap[name]
		if ok {
			return errors.New(fmt.Sprintf("pipeline exists:%s", name))
		}

		pipeline.name = name
		p, err := initPipeline(pipeline, c, c.application)
		if err != nil {
			return err
		}
		c.pipelineRawContent[name] = struct{ *Pipeline }{Pipeline: pipeline}
		c.pipelineMap[name] = p
	}

	return nil
}
