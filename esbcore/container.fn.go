package esbcore

import (
	"errors"
	"strings"

	"esbconcept/esbapi"
)

func (c *ContainerInst) RegisterBuiltinFn(methodName string, fg esbapi.FnGen) error {
	if !strings.HasPrefix(methodName, "@") {
		return errors.New("builtin functions should have @ as prefix")
	}
	_, ok := c.builtinGenFnMap[methodName]
	if ok {
		return errors.New("method already registered:" + methodName)
	}
	c.builtinGenFnMap[methodName] = fg
	return nil
}

func (c *ContainerInst) RegisterCustomFn(name string, fn esbapi.FnGen) error {
	if !strings.HasPrefix(name, "#") {
		return errors.New("custom functions should have # as prefix")
	}
	_, ok := c.customGenFnMap[name]
	if ok {
		return errors.New("custom function already exists:" + name)
	}
	c.customGenFnMap[name] = fn
	return nil
}

func (c *ContainerInst) RegisterSourceConnectorGen(name string, connGen esbapi.SourceConnectorGenerator) error {
	if _, ok := c.registerSourceConnectorGen[name]; ok {
		return errors.New("source connector generator already exists:" + name)
	}
	c.registerSourceConnectorGen[name] = connGen
	return nil
}

func (c *ContainerInst) RegisterTargetConnectorGen(name string, connGen esbapi.TargetConnectorGenerator) error {
	if _, ok := c.registerTargetConnectorGen[name]; ok {
		return errors.New("target connector generator already exists:" + name)
	}
	c.registerTargetConnectorGen[name] = connGen
	return nil
}
