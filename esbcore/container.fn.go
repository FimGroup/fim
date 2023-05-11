package esbcore

import (
	"errors"

	"esbconcept/esbapi"
)

func (c *Container) RegisterBuiltinFn(methodName string, fg esbapi.FnGen) error {
	_, ok := c.builtinGenFnMap[methodName]
	if ok {
		return errors.New("method already registered:" + methodName)
	}
	c.builtinGenFnMap[methodName] = fg
	return nil
}
