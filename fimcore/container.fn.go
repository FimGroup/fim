package fimcore

import (
	"errors"
	"strings"

	"github.com/ThisIsSun/fim/fimapi/pluginapi"
)

func (c *ContainerInst) RegisterBuiltinFn(methodName string, fg pluginapi.FnGen) error {
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

func (c *ContainerInst) RegisterCustomFn(name string, fn pluginapi.FnGen) error {
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

func (c *ContainerInst) RegisterSourceConnectorGen(connGen pluginapi.SourceConnectorGenerator) error {
	names := connGen.GeneratorNames()
	for _, name := range names {
		if _, ok := c.registerSourceConnectorGen[name]; ok {
			return errors.New("source connector generator already exists:" + name)
		}
	}
	for _, name := range names {
		c.registerSourceConnectorGen[name] = connGen
	}
	return nil
}

func (c *ContainerInst) RegisterTargetConnectorGen(connGen pluginapi.TargetConnectorGenerator) error {
	names := connGen.GeneratorNames()
	for _, name := range names {
		if !strings.HasPrefix(name, "&") {
			name = "&" + name
		}
		if _, ok := c.registerTargetConnectorGen[name]; ok {
			return errors.New("target connector generator already exists:" + name)
		}
	}
	for _, name := range names {
		if !strings.HasPrefix(name, "&") {
			name = "&" + name
		}
		c.registerTargetConnectorGen[name] = connGen
	}
	return nil
}
