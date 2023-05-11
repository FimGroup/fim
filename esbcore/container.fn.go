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
