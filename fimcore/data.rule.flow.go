package fimcore

import (
	"errors"
	"reflect"
	"strings"

	"github.com/FimGroup/fim/fimapi/basicapi"
	"github.com/FimGroup/fim/fimapi/pluginapi"
	"github.com/FimGroup/fim/fimapi/rule"
)

type templateFlow struct {
	In     [][]string                            `toml:"in"`
	Out    [][]string                            `toml:"out"`
	PreOut [][]string                            `toml:"pre_out"`
	Flow   map[string][]map[string][]interface{} `toml:"flow"`
}

type Flow struct {
	dtd       *DataTypeDefinitions
	container *ContainerInst

	inParamList []struct {
		ModelFieldPath string
		SplitPath      []string
		DataType       pluginapi.DataType
		KeySplitPath   []string
		KeyPath        string
	}
	outParamList []struct {
		ModelFieldPath string
		SplitPath      []string
		DataType       pluginapi.DataType
		KeySplitPath   []string
		KeyPath        string
	}
	localPreOutOperations map[string]struct {
		Operation string
		SplitPath []string
	}

	fnList []pluginapi.Fn
}

func NewFlow(dtd *DataTypeDefinitions, c *ContainerInst) *Flow {
	return &Flow{
		dtd:       dtd,
		container: c,

		inParamList: []struct {
			ModelFieldPath string
			SplitPath      []string
			DataType       pluginapi.DataType
			KeySplitPath   []string
			KeyPath        string
		}{},
		outParamList: []struct {
			ModelFieldPath string
			SplitPath      []string
			DataType       pluginapi.DataType
			KeySplitPath   []string
			KeyPath        string
		}{},
		localPreOutOperations: map[string]struct {
			Operation string
			SplitPath []string
		}{},
	}
}

func (f *Flow) mergeToml(tf *templateFlow) error {
	for _, paramPair := range tf.In {
		if len(paramPair) != 2 {
			return errors.New("parameter pair should have 2 items")
		}
		// path -> local
		if err := f.addIn(paramPair[0], paramPair[1]); err != nil {
			return err
		}
	}
	for _, paramPair := range tf.Out {
		if len(paramPair) != 2 {
			return errors.New("parameter pair should have 2 items")
		}
		// local -> path
		if err := f.addOut(paramPair[0], paramPair[1]); err != nil {
			return err
		}
	}
	for _, paramPair := range tf.PreOut {
		if len(paramPair) != 2 {
			return errors.New("parameter pair should have 2 items")
		}
		// operation -> path
		if err := f.addPreOut(paramPair[0], paramPair[1]); err != nil {
			return err
		}
	}
	if err := f.validateLocalParameters(); err != nil {
		return err
	}
	if err := f.addFlow(tf); err != nil {
		return err
	}

	return nil
}

func (f *Flow) addIn(source, local string) error {
	if !rule.ValidateFullPath(source) {
		return errors.New("in parameter path invalid:" + source)
	}

	if dt, _, err := f.dtd.TypeOfPath(source); err != nil {
		return err
	} else if dt == pluginapi.DataTypeUnavailable {
		return errors.New("cannot find path:" + source)
	} else {
		f.inParamList = append(f.inParamList, struct {
			ModelFieldPath string
			SplitPath      []string
			DataType       pluginapi.DataType
			KeySplitPath   []string
			KeyPath        string
		}{
			ModelFieldPath: source,
			SplitPath:      rule.SplitFullPath(source),
			DataType:       dt,
			KeySplitPath:   rule.SplitFullPath(local),
			KeyPath:        local,
		})
	}

	return nil
}

func (f *Flow) inConv() func(source, local *ModelInst) error {
	return func(source, local *ModelInst) error {
		for _, dStruct := range f.inParamList {
			if err := source.transferTo(local, dStruct.SplitPath, dStruct.KeySplitPath, ByLeft); err != nil {
				return err
			}
		}
		return nil
	}
}

func (f *Flow) addOut(local, out string) error {
	if !rule.ValidateFullPath(out) {
		return errors.New("out parameter path invalid:" + out)
	}

	if dt, _, err := f.dtd.TypeOfPath(out); err != nil {
		return err
	} else if dt == pluginapi.DataTypeUnavailable {
		return errors.New("cannot find path:" + out)
	} else {
		f.outParamList = append(f.outParamList, struct {
			ModelFieldPath string
			SplitPath      []string
			DataType       pluginapi.DataType
			KeySplitPath   []string
			KeyPath        string
		}{
			ModelFieldPath: out,
			SplitPath:      rule.SplitFullPath(out),
			DataType:       dt,
			KeySplitPath:   rule.SplitFullPath(local),
			KeyPath:        local,
		})
	}

	return nil
}

func (f *Flow) outConv() func(local, out *ModelInst) error {
	return func(local, out *ModelInst) error {
		// process pre_out
		for _, op := range f.localPreOutOperations {
			switch op.Operation {
			case "@remove":
				if err := out.deleteField(op.SplitPath); err != nil {
					return err
				}
			default:
				return errors.New("unknown pre_out operation:" + op.Operation)
			}
		}

		// process out
		for _, dStruct := range f.outParamList {
			if err := local.transferTo(out, dStruct.KeySplitPath, dStruct.SplitPath, ByRight); err != nil {
				return err
			}
		}

		return nil
	}
}

func (f *Flow) FlowFn(casePreFn func(m pluginapi.Model) (bool, error)) func() func(global pluginapi.Model) error {
	return func() func(global pluginapi.Model) error {
		local := NewModelInst(f.dtd)
		return func(global pluginapi.Model) error {
			if casePreFn != nil {
				val, err := casePreFn(global)
				if err != nil {
					return err
				}
				if !val {
					return nil
				}
			}

			if err := f.inConv()(global.(*ModelInst), local); err != nil {
				return err
			}
			// process flow
			{
				for _, fn := range f.fnList {
					if err := fn(local); err != nil {
						return err
					}
				}
			}
			if err := f.outConv()(local, global.(*ModelInst)); err != nil {
				return err
			}

			return nil
		}
	}
}

func (f *Flow) FlowFnNoResp(casePreFn func(m pluginapi.Model) (bool, error)) func() func(global pluginapi.Model) error {
	return func() func(global pluginapi.Model) error {
		local := NewModelInst(f.dtd)
		return func(global pluginapi.Model) error {
			if casePreFn != nil {
				val, err := casePreFn(global)
				if err != nil {
					return err
				}
				if !val {
					return nil
				}
			}

			if err := f.inConv()(global.(*ModelInst), local); err != nil {
				return err
			}
			// process flow
			{
				for _, fn := range f.fnList {
					if err := fn(local); err != nil {
						return err
					}
				}
			}
			dummy := NewModelInst(f.dtd)
			if err := f.outConv()(local, dummy); err != nil {
				return err
			}

			return nil
		}
	}
}

func (f *Flow) addFlow(tf *templateFlow) error {
	steps := tf.Flow["steps"]
	var fList []pluginapi.Fn
	for _, step := range steps {
		var concreteFn pluginapi.Fn
		var wrapperFn func(fn pluginapi.Fn) pluginapi.Fn
		for fn, params := range step {
			// to make sure every step struct only contains one step, so overwrite may happen when duplicated definition
			if strings.HasPrefix(fn, "@case-") {
				f, err := f.prepareCaseClause(fn, params)
				if err != nil {
					return err
				}
				wrapperFn = f
			} else if fn[0] == '@' {
				//builtin function
				fngen, ok := f.container.builtinGenFnMap[fn]
				if !ok {
					return errors.New("builtin function not found:" + fn)
				}
				fnInst, err := fngen(params)
				if err != nil {
					return err
				}
				concreteFn = fnInst
			} else if fn[0] == '#' {
				//user defined function
				fngen, ok := f.container.customGenFnMap[fn]
				if !ok {
					return errors.New("user defined function not found:" + fn)
				}
				fnInst, err := fngen(params)
				if err != nil {
					return err
				}
				concreteFn = fnInst
			} else {
				return errors.New("unknown command:" + fn)
			}
		}
		if concreteFn == nil {
			return errors.New("step does not contains any logic")
		}
		if wrapperFn != nil {
			concreteFn = wrapperFn(concreteFn)
		}
		fList = append(fList, concreteFn)
	}
	f.fnList = fList
	return nil
}

func (f *Flow) validateLocalParameters() error {
	outLocalMapping := map[string]string{}
	for _, v := range f.outParamList {
		outLocalMapping[v.KeyPath] = v.ModelFieldPath
	}
	// compare the types of the same local parameters
	for _, si := range f.inParamList {
		modelFieldPath, ok := outLocalMapping[si.KeyPath]
		if ok {
			sdt, spdt, err := f.dtd.TypeOfPath(si.ModelFieldPath)
			if err != nil {
				return err
			}
			ddt, dpdt, err := f.dtd.TypeOfPath(modelFieldPath)
			if err != nil {
				return err
			}
			if sdt != ddt || spdt != dpdt {
				return errors.New("local parameter types of in and out do not match:" + si.KeyPath)
			}
		}
	}
	return nil
}

func (f *Flow) addPreOut(op string, path string) error {
	_, ok := f.localPreOutOperations[path]
	if ok {
		return errors.New("path registered:" + path)
	}

	if !rule.ValidateFullPath(path) {
		return errors.New("path invalid:" + path)
	}

	if dt, _, err := f.dtd.TypeOfPath(path); err != nil {
		return err
	} else if dt == pluginapi.DataTypeUnavailable {
		return errors.New("cannot find path:" + path)
	} else {
		f.localPreOutOperations[path] = struct {
			Operation string
			SplitPath []string
		}{
			Operation: op,
			SplitPath: rule.SplitFullPath(path),
		}
	}

	return nil
}

func (f *Flow) prepareCaseClause(fn string, params []interface{}) (func(fn pluginapi.Fn) pluginapi.Fn, error) {
	if len(params) <= 0 {
		return nil, errors.New("case clause requires more parameters:" + fn)
	}
	path, ok := params[0].(string)
	if !ok {
		return nil, errors.New("case clause parameter should be a string:" + fn)
	}
	paths := rule.SplitFullPath(path)
	switch fn {
	case "@case-true":
		return func(fn pluginapi.Fn) pluginapi.Fn {
			return func(m basicapi.Model) error {
				val := m.GetFieldUnsafe(paths)
				if val == nil {
					return errors.New("case-true on nil value:" + path)
				}
				b, ok := val.(bool)
				if !ok {
					return errors.New("case-true on non-bool value:" + path)
				}
				if b {
					return fn(m)
				} else {
					return nil
				}
			}
		}, nil
	case "@case-false":
		return func(fn pluginapi.Fn) pluginapi.Fn {
			return func(m basicapi.Model) error {
				val := m.GetFieldUnsafe(paths)
				if val == nil {
					return errors.New("case-true on nil value:" + path)
				}
				b, ok := val.(bool)
				if !ok {
					return errors.New("case-true on non-bool value:" + path)
				}
				if !b {
					return fn(m)
				} else {
					return nil
				}
			}
		}, nil
	case "@case-equals":
		if len(params) != 2 {
			return nil, errors.New("@case-equals requires 2 parameters")
		}
		path2, ok := params[1].(string)
		if !ok {
			return nil, errors.New("case clause parameter should be a string:" + fn)
		}
		paths2 := rule.SplitFullPath(path2)
		return func(fn pluginapi.Fn) pluginapi.Fn {
			return func(m basicapi.Model) error {
				val := m.GetFieldUnsafe(paths)
				val2 := m.GetFieldUnsafe(paths2)
				if compareInterfaceValue(val, val2) {
					return fn(m)
				} else {
					return nil
				}
			}
		}, nil
	case "@case-not-equals":
		if len(params) != 2 {
			return nil, errors.New("@case-equals requires 2 parameters")
		}
		path2, ok := params[1].(string)
		if !ok {
			return nil, errors.New("case clause parameter should be a string:" + fn)
		}
		paths2 := rule.SplitFullPath(path2)
		return func(fn pluginapi.Fn) pluginapi.Fn {
			return func(m basicapi.Model) error {
				val := m.GetFieldUnsafe(paths)
				val2 := m.GetFieldUnsafe(paths2)
				if !compareInterfaceValue(val, val2) {
					return fn(m)
				} else {
					return nil
				}
			}
		}, nil
	case "@case-empty":
		return func(fn pluginapi.Fn) pluginapi.Fn {
			return func(m basicapi.Model) error {
				val := m.GetFieldUnsafe(paths)
				if val == nil {
					return fn(m)
				}
				v, ok := val.(string)
				if !ok {
					return errors.New("case-empty on non-string value:" + path)
				}
				if v == "" {
					return fn(m)
				} else {
					return nil
				}
			}
		}, nil
	case "@case-non-empty":
		return func(fn pluginapi.Fn) pluginapi.Fn {
			return func(m basicapi.Model) error {
				val := m.GetFieldUnsafe(paths)
				if val == nil {
					return nil
				}
				v, ok := val.(string)
				if !ok {
					return errors.New("case-non-empty on non-string value:" + path)
				}
				if v != "" {
					return fn(m)
				} else {
					return nil
				}
			}
		}, nil
	default:
		return nil, errors.New("unknown case clause:" + fn)
	}
}

func compareInterfaceValue(val interface{}, val2 interface{}) bool {
	// for int type, convert all to int64
	//FIXME need better solution
	isValInt := false
	isVal2Int := false
	switch val.(type) {
	case int:
		isValInt = true
	case int8:
		isValInt = true
	case int16:
		isValInt = true
	case int32:
		isValInt = true
	case int64:
		isValInt = true
	}
	switch val2.(type) {
	case int:
		isVal2Int = true
	case int8:
		isVal2Int = true
	case int16:
		isVal2Int = true
	case int32:
		isVal2Int = true
	case int64:
		isVal2Int = true
	}
	if isValInt && isVal2Int {
		return reflect.ValueOf(val).Int() == reflect.ValueOf(val2).Int()
	}

	// for other types, remain the same
	//FIXME for float32/float64, still require conversion
	return val == val2
}
