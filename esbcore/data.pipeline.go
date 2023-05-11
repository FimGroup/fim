package esbcore

import (
	"bytes"
	"errors"
	"log"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

type PipelineProcess func(m *ModelInst) error

type ConnectorDataMapping map[string]string

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
		In  ConnectorDataMapping `toml:"in"`
		Out ConnectorDataMapping `toml:"out"`
	} `toml:"connector_mapping"`

	container          *ContainerInst
	connectorInitFuncs []func(PipelineProcess) error
	steps              []func() func(global *ModelInst) error
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
			gen, ok := sourceConnectorGenMap[connectorName]
			if !ok {
				return nil, errors.New("source connector generator cannot be found:" + connectorName)
			}
			if f, err := gen(v, p.container); err != nil {
				return nil, err
			} else {
				p.connectorInitFuncs = append(p.connectorInitFuncs, f)
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
				g, ok := targetConnectorGenMap[flow]
				if !ok {
					return nil, errors.New("target connector cannot be found:" + flow)
				}
				f, err := g(v)
				if err != nil {
					return nil, err
				}
				if okS {
					p.steps = append(p.steps, func() func(g *ModelInst) error {
						return func(g *ModelInst) error {
							return f(g, g)
						}
					})
				} else {
					p.steps = append(p.steps, func() func(g *ModelInst) error {
						return func(g *ModelInst) error {
							return f(g, NewModelInst(p.container.flowModel))
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

func (p *Pipeline) toPipelineFn() PipelineProcess {
	return func(m *ModelInst) error {
		for _, fn := range p.steps {
			f := fn()
			if err := f(m); err != nil {
				return err
			}
		}
		return nil
	}
}

func (p *Pipeline) RunPipeline() error {
	// start source connector
	process := p.toPipelineFn()
	for _, f := range p.connectorInitFuncs {
		go func(fn func(pipelineProcess PipelineProcess) error) {
			if err := fn(process); err != nil {
				log.Fatalln(err)
			}
		}(f)
	}
	log.Println("RunPipeline done.")
	return nil
}
