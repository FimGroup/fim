package fimcore

import (
	"bytes"
	"errors"
	"log"
	"strings"

	"github.com/pelletier/go-toml/v2"

	"github.com/ThisIsSun/fim/fimapi/pluginapi"
)

type Pipeline struct {
	Parameter struct {
		Inputs     []string            `toml:"inputs"`
		PreOutputs []map[string]string `toml:"pre_outputs"`
		Outputs    []string            `toml:"outputs"`
	} `toml:"parameter"`
	Pipeline struct {
		Steps            []map[string]string `toml:"steps"`
		SourceConnectors []map[string]string `toml:"source_connectors"`
	} `toml:"pipeline"`
	ConnectorMapping map[string]struct {
		Req pluginapi.DataMapping `toml:"req"`
		Res pluginapi.DataMapping `toml:"res"`
	} `toml:"connector_mapping"`

	container          *ContainerInst
	connectorInitFuncs []struct {
		pluginapi.ConnectorProcessEntryPoint
		*pluginapi.MappingDefinition
	}
	steps []func() func(global pluginapi.Model) error
}

func NewPipeline(tomlContent string, container *ContainerInst) (*Pipeline, error) {
	p := new(Pipeline)
	if err := toml.NewDecoder(bytes.NewBufferString(tomlContent)).DisallowUnknownFields().Decode(p); err != nil {
		return nil, err
	}
	p.container = container

	// parse pipeline definition and validate components
	// 1. validate parameter
	// currently not supported
	// 2. validate pipeline.source_connectors
	{
		for _, v := range p.Pipeline.SourceConnectors {
			connectorName, ok := v["@connector"]
			if !ok {
				return nil, errors.New("no @connector defined")
			}

			// connector mapping
			connInstName, ok := v["@mapping"]
			if !ok {
				return nil, errors.New("no @mapping defined")
			}
			s, ok := p.ConnectorMapping[connInstName]
			if !ok {
				return nil, errors.New("connect mapping cannot be found:" + connInstName)
			}
			mappdingDef := &pluginapi.MappingDefinition{
				Req: s.Req,
				Res: s.Res,
			}

			gen, ok := container.registerSourceConnectorGen[connectorName]
			if !ok {
				return nil, errors.New("source connector generator cannot be found:" + connectorName)
			}
			if f, err := gen(v, p.container); err != nil {
				return nil, err
			} else {
				container.connectorMap[f.InstanceName] = f.Connector
				p.connectorInitFuncs = append(p.connectorInitFuncs, struct {
					pluginapi.ConnectorProcessEntryPoint
					*pluginapi.MappingDefinition
				}{ConnectorProcessEntryPoint: f.ConnectorProcessEntryPoint, MappingDefinition: mappdingDef})
			}
		}
	}
	// 3. validate pipeline.steps
	{
		for _, v := range p.Pipeline.Steps {
			flowS, okS := v["@flow"]
			flowA, okA := v["#flow"]
			var flow string
			if okS && okA {
				return nil, errors.New("should not make a pipeline step both invoking flow and triggering event step")
			} else if okS {
				flow = flowS
			} else if okA {
				flow = flowA
			} else {
				return nil, errors.New("no flow name defined in step")
			}

			if strings.HasPrefix(flow, "&") {
				// target connector
				g, ok := container.registerTargetConnectorGen[flow]
				if !ok {
					return nil, errors.New("target connector cannot be found:" + flow)
				}

				// connector mapping
				connInstName, ok := v["@mapping"]
				if !ok {
					return nil, errors.New("no @mapping defined")
				}
				s, ok := p.ConnectorMapping[connInstName]
				if !ok {
					return nil, errors.New("connect mapping cannot be found:" + connInstName)
				}
				mappdingDef := &pluginapi.MappingDefinition{
					Req: s.Req,
					Res: s.Res,
				}
				//FIXME support parameter data mapping for target connector

				tConnStruct, err := g(v, container, mappdingDef)
				if err != nil {
					return nil, err
				}
				flowInst := tConnStruct.ConnectorFlow
				if _, ok := container.connectorMap[tConnStruct.InstanceName]; !ok {
					// add connector lifecycle map if new
					container.connectorMap[tConnStruct.InstanceName] = tConnStruct.Connector
				}
				// assemble flow
				if okS {
					p.steps = append(p.steps, func() func(g pluginapi.Model) error {
						return func(g pluginapi.Model) error {
							return flowInst(g, g)
						}
					})
				} else {
					p.steps = append(p.steps, func() func(g pluginapi.Model) error {
						return func(g pluginapi.Model) error {
							return flowInst(g, NewModelInst(p.container.flowModel))
						}
					})
				}
			} else {
				// flow
				f, ok := p.container.flowMap[flow]
				if !ok {
					return nil, errors.New("flow cannot be found:" + flow)
				}

				if okS {
					// sync/invoke flow
					p.steps = append(p.steps, f.FlowFn)
				} else {
					// async/trigger event
					p.steps = append(p.steps, f.FlowFnNoResp)
				}
			}
		}
	}

	return p, nil
}

func (p *Pipeline) toPipelineFn() pluginapi.PipelineProcess {
	return func(m pluginapi.Model) error {
		for _, fn := range p.steps {
			f := fn()
			if err := f(m); err != nil {
				return err
			}
		}
		return nil
	}
}

func (p *Pipeline) setupPipeline() error {
	// start source connector
	process := p.toPipelineFn()
	for _, f := range p.connectorInitFuncs {
		if err := f.ConnectorProcessEntryPoint(process, f.MappingDefinition); err != nil {
			return err
		}
	}

	log.Println("setupPipeline done.")
	return nil
}
