package core

import "github.com/FimGroup/fim/fim-ng/api/definition"

type Engine struct {
}

func NewEngine() *Engine {
	return &Engine{}
}

func (e *Engine) AddPathBuilder(p definition.PathBuilder) {
	//TODO
	p.Configure()
}
