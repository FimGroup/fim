package fimcore

import (
	"errors"
	"log"
	"strings"

	"github.com/FimGroup/fim/fimapi/pluginapi"
	"github.com/FimGroup/fim/fimapi/rule"
	"github.com/FimGroup/fim/fimcore/modelinst"
)

type Pipeline struct {
	Metadata struct {
		Version string `toml:"version"`
	} `toml:"metadata"`
	Parameter struct {
		Inputs        []string            `toml:"inputs"`
		PreOutputs    []map[string]string `toml:"pre_outputs"`
		Outputs       []string            `toml:"outputs"`
		LocalVariable []map[string]string `toml:"local_variables"`
	} `toml:"parameter"`
	Pipeline struct {
		Steps            [][][]interface{} `toml:"steps"`
		SourceConnectors [][][]interface{} `toml:"source_connectors"`
	} `toml:"pipeline"`
	ConnectorMapping map[string]struct {
		Req       modelinst.MappingRuleRaw `toml:"req"`
		Res       modelinst.MappingRuleRaw `toml:"res"`
		ErrSimple []map[string]string      `toml:"err_simple"`
	} `toml:"connector_mapping"`

	container          *ContainerInst
	connectorInitFuncs []struct {
		pluginapi.SourceConnector
		*pluginapi.MappingDefinition
	}
	steps []func() func(global pluginapi.Model) error
}

func initPipeline(p *Pipeline, container *ContainerInst) (*Pipeline, error) {
	p.container = container

	if len(p.Metadata.Version) == 0 {
		return nil, errors.New("pipeline version is empty")
	}
	// parse pipeline definition and validate components
	// 1. validate parameter
	// currently not supported
	// 2. validate pipeline.source_connectors
	{
		// build source connector maps
		var sourceConnectorMapLists []map[string]string
		for _, v := range p.Pipeline.SourceConnectors {
			m := map[string]string{}
			for _, vv := range v {
				if len(vv) != 2 {
					return nil, errors.New("not k-v pair in source connector definition")
				}
				var k, v string
				if sv, ok := vv[0].(string); !ok {
					return nil, errors.New("not string key in source connector pair")
				} else {
					k = sv
				}
				if sv, ok := vv[1].(string); !ok {
					return nil, errors.New("not string value in source connector pair")
				} else {
					v = container.configureManager.ReplaceStaticConfigure(sv)
				}
				if _, ok := m[k]; ok {
					return nil, errors.New("duplicated key in source connector definition")
				} else {
					m[k] = v
				}
			}
			sourceConnectorMapLists = append(sourceConnectorMapLists, m)
		}
		// do source connector
		for _, v := range sourceConnectorMapLists {
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
			resConverter, err := s.Res.ToConverter()
			if err != nil {
				return nil, err
			}
			reqConverter, err := s.Req.ToConverter()
			if err != nil {
				return nil, err
			}
			mappdingDef := &pluginapi.MappingDefinition{
				ReqConverter: reqConverter.GeneralTransfer,
				ReqArgPaths:  reqConverter.TargetLeafPathList,
				ResConverter: resConverter.GeneralTransfer,
				ResArgPaths:  resConverter.SourceLeafPathList,
				ErrSimple:    s.ErrSimple,
			}

			gen, ok := container.registerSourceConnectorGen[connectorName]
			if !ok {
				return nil, errors.New("source connector generator cannot be found:" + connectorName)
			}
			if f, err := gen.GenerateSourceConnectorInstance(v, p.container); err != nil {
				return nil, err
			} else {
				container.connectorMap[f.ConnectorName()] = f
				p.connectorInitFuncs = append(p.connectorInitFuncs, struct {
					pluginapi.SourceConnector
					*pluginapi.MappingDefinition
				}{SourceConnector: f, MappingDefinition: mappdingDef})
			}
		}
	}
	// 3. validate pipeline.steps
	{
		// build pipeline.steps maps
		var stepsMapList []map[string]string
		for _, v := range p.Pipeline.Steps {
			m := map[string]string{}
			for _, vv := range v {
				if len(vv) != 2 {
					return nil, errors.New("not k-v pair in pipeline.steps definition")
				}
				var k, v string
				if sv, ok := vv[0].(string); !ok {
					return nil, errors.New("not string key in pipeline.steps pair")
				} else {
					k = sv
				}
				if sv, ok := vv[1].(string); !ok {
					return nil, errors.New("not string value in pipeline.steps pair")
				} else {
					v = container.configureManager.ReplaceStaticConfigure(sv)
				}
				if _, ok := m[k]; ok {
					return nil, errors.New("duplicated key in pipeline.steps definition")
				} else {
					m[k] = v
				}
			}
			stepsMapList = append(stepsMapList, m)
		}
		// do pipeline.steps
		for _, v := range stepsMapList {
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

			// process case clause
			var casePreFn func(m pluginapi.Model) (bool, error)
			caseOperator, caseValue := findCaseClause(v)
			if caseOperator != "" {
				var err error
				casePreFn, err = generateCasePreFn(caseOperator, caseValue)
				if err != nil {
					return nil, err
				}
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
				resConverter, err := s.Res.ToConverter()
				if err != nil {
					return nil, err
				}
				reqConverter, err := s.Req.ToConverter()
				if err != nil {
					return nil, err
				}
				mappdingDef := &pluginapi.MappingDefinition{
					ReqConverter: reqConverter.GeneralTransfer,
					ReqArgPaths:  reqConverter.TargetLeafPathList,
					ResConverter: resConverter.GeneralTransfer,
					ResArgPaths:  resConverter.SourceLeafPathList,
					ErrSimple:    []map[string]string{},
				}
				//FIXME support parameter data mapping for target connector

				tConnector, err := g.GenerateTargetConnectorInstance(v, container, mappdingDef)
				if err != nil {
					return nil, err
				}
				flowInst := tConnector.InvokeFlow
				if _, ok := container.connectorMap[tConnector.ConnectorName()]; !ok {
					// add connector lifecycle map if new
					container.connectorMap[tConnector.ConnectorName()] = tConnector
				}
				// assemble flow
				if okS {
					p.steps = append(p.steps, func() func(g pluginapi.Model) error {
						return func(g pluginapi.Model) error {
							if casePreFn != nil {
								match, err := casePreFn(g)
								if err != nil {
									return err
								}
								if !match {
									return nil
								}
							}
							return flowInst(g, g)
						}
					})
				} else {
					p.steps = append(p.steps, func() func(g pluginapi.Model) error {
						return func(g pluginapi.Model) error {
							if casePreFn != nil {
								match, err := casePreFn(g)
								if err != nil {
									return err
								}
								if !match {
									return nil
								}
							}
							return flowInst(g, p.container.NewModel())
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
					p.steps = append(p.steps, f.FlowFn(casePreFn))
				} else {
					// async/trigger event
					p.steps = append(p.steps, f.FlowFnNoResp(casePreFn))
				}
			}
		}
	}

	return p, nil
}

func generateCasePreFn(operator string, value string) (func(m pluginapi.Model) (bool, error), error) {
	paths := rule.SplitFullPath(value)
	switch operator {
	case "@case-true":
		return func(m pluginapi.Model) (bool, error) {
			val := m.GetFieldUnsafe0(paths)
			if val == nil {
				return false, errors.New("case-true on nil value:" + value)
			}
			b, ok := val.(bool)
			if !ok {
				return false, errors.New("case-true on non-bool value:" + value)
			}
			return b, nil
		}, nil
	case "@case-false":
		return func(m pluginapi.Model) (bool, error) {
			val := m.GetFieldUnsafe0(paths)
			if val == nil {
				return false, errors.New("case-false on nil value:" + value)
			}
			b, ok := val.(bool)
			if !ok {
				return false, errors.New("case-false on non-bool value:" + value)
			}
			return !b, nil
		}, nil
	case "@case-empty":
		return func(m pluginapi.Model) (bool, error) {
			val := m.GetFieldUnsafe0(paths)
			if val == nil {
				return true, nil
			}
			v, ok := val.(string)
			if !ok {
				return false, errors.New("case-empty on non-string value:" + value)
			}
			if v == "" {
				return true, nil
			} else {
				return false, nil
			}
		}, nil
	case "@case-non-empty":
		return func(m pluginapi.Model) (bool, error) {
			val := m.GetFieldUnsafe0(paths)
			if val == nil {
				return false, nil
			}
			v, ok := val.(string)
			if !ok {
				return false, errors.New("case-non-empty on non-string value:" + value)
			}
			if v != "" {
				return true, nil
			} else {
				return false, nil
			}
		}, nil
	default:
		return nil, errors.New("unknown case clause:" + operator)
	}
}

func findCaseClause(v map[string]string) (string, string) {
	for op, val := range v {
		if strings.HasPrefix(op, "@case-") {
			return op, val
		}
	}
	return "", ""
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
		if err := f.InvokeProcess(process, f.MappingDefinition); err != nil {
			return err
		}
	}

	log.Println("setupPipeline done.")
	return nil
}
