package esbcore

type Fn func(m *ModelInst) error

var builtinGenFnMap map[string]func(params []interface{}) (Fn, error)

func init() {
	builtinGenFnMap = map[string]func(params []interface{}) (Fn, error){}
	builtinGenFnMap["@assign"] = FnAssign
}

func FnAssign(params []interface{}) (Fn, error) {
	var field string = params[0].(string)
	var val = params[1]
	return func(m *ModelInst) error {
		return m.addOrUpdateField(field, val)
	}, nil
}
