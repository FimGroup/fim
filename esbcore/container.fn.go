package esbcore

import "errors"

type Fn func(m *ModelInst) error

func (c *Container) RegisterBuiltinFn(methodName string, fg func(params []interface{}) (Fn, error)) error {
	_, ok := c.builtinGenFnMap[methodName]
	if ok {
		return errors.New("method already registered:" + methodName)
	}
	c.builtinGenFnMap[methodName] = fg
	return nil
}
