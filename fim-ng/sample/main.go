package main

import (
	"github.com/FimGroup/fim/fim-ng/api/definition"
	"github.com/FimGroup/fim/fim-ng/core"
)

type SamplePath struct {
	definition.PathDef
}

func (p *SamplePath) Configure() {
	p.NewPath(definition.PathSettings{}).
		Next().
		Next().
		Next()
}

func main() {
	engine := core.NewEngine()

	engine.AddPathBuilder(new(SamplePath))
}
