package fimcore

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/FimGroup/fim/fimapi/basicapi"
	"github.com/FimGroup/fim/fimapi/pluginapi"
	"github.com/FimGroup/fim/fimapi/rule"
	"github.com/FimGroup/fim/fimcore/modelinst"
)

type templateFlow struct {
	In     modelinst.MappingRuleRaw              `toml:"in"`
	Out    modelinst.MappingRuleRaw              `toml:"out"`
	PreOut [][]string                            `toml:"pre_out"`
	Flow   map[string][]map[string][]interface{} `toml:"flow"`
}

type Flow struct {
	dtd       *DataTypeDefinitions
	container *ContainerInst

	inConverter           *modelinst.ModelConverter
	outConverter          *modelinst.ModelConverter
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

		localPreOutOperations: map[string]struct {
			Operation string
			SplitPath []string
		}{},
	}
}

func (f *Flow) mergeToml(tf *templateFlow) error {

	if inConverter, err := tf.In.ToConverter(); err != nil {
		return err
	} else if err := f.checkInDtd(inConverter.SourceLeafPathList); err != nil {
		return err
	} else {
		f.inConverter = inConverter
	}
	if outConverter, err := tf.Out.ToConverter(); err != nil {
		return err
	} else if err := f.checkInDtd(outConverter.TargetLeafPathList); err != nil {
		return err
	} else {
		f.outConverter = outConverter
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
	// validation
	if err := f.validateRule(); err != nil {
		return err
	}
	if err := f.addFlow(tf); err != nil {
		return err
	}

	return nil
}

func (f *Flow) validateRule() error {
	// check in/out data type
	inMap := map[string]string{}
	for idx, key := range f.inConverter.TargetLeafPathList {
		inMap[key] = f.inConverter.SourceLeafPathList[idx]
	}
	outMap := map[string]string{}
	for idx, key := range f.outConverter.SourceLeafPathList {
		outMap[key] = f.outConverter.TargetLeafPathList[idx]
	}
	for key, path := range inMap {
		oPath, ok := outMap[key]
		if !ok {
			continue
		}
		sdt, _, err := f.dtd.TypeOfPath(path)
		if err != nil {
			return err
		}
		ddt, _, err := f.dtd.TypeOfPath(oPath)
		if err != nil {
			return err
		}
		if sdt != ddt {
			return errors.New(fmt.Sprintf("flow parameter=[%s] input and output mapping types are not the same", key))
		}
	}
	return nil
}

func (f *Flow) checkInDtd(paths []string) error {
	for _, path := range paths {
		if !rule.ValidateFullPathOfDefinition(path) {
			return errors.New("parameter path invalid:" + path)
		}
		if dt, _, err := f.dtd.TypeOfPath(path); err != nil {
			return err
		} else if dt == pluginapi.DataTypeUnavailable {
			return errors.New("cannot find path:" + path)
		}
	}
	return nil
}

func (f *Flow) inConv() func(source, local modelinst.ModelInst2) error {
	return func(source, local modelinst.ModelInst2) error {
		return f.inConverter.Transfer(source, local)
	}
}

func (f *Flow) outConv() func(local, out modelinst.ModelInst2) error {
	return func(local, out modelinst.ModelInst2) error {
		// process pre_out
		for _, op := range f.localPreOutOperations {
			switch op.Operation {
			case "@remove-object":
				if err := out.RemoveObjectByPath(op.SplitPath); err != nil {
					return err
				}
			default:
				return errors.New("unknown pre_out operation:" + op.Operation)
			}
		}

		// process out
		return f.outConverter.Transfer(local, out)
	}
}

func (f *Flow) FlowFn(casePreFn func(m pluginapi.Model) (bool, error)) func() func(global pluginapi.Model) error {
	return func() func(global pluginapi.Model) error {
		local := modelinst.ModelInstHelper{}.NewInst()
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

			if err := f.inConv()(global.(modelinst.ModelInst2), local); err != nil {
				return err
			}
			// process flow
			{
				for _, fn := range f.fnList {
					if err := fn(local.(pluginapi.Model)); err != nil {
						return err
					}
				}
			}
			if err := f.outConv()(local, global.(modelinst.ModelInst2)); err != nil {
				return err
			}

			return nil
		}
	}
}

func (f *Flow) FlowFnNoResp(casePreFn func(m pluginapi.Model) (bool, error)) func() func(global pluginapi.Model) error {
	return func() func(global pluginapi.Model) error {
		local := modelinst.ModelInstHelper{}.NewInst()
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

			if err := f.inConv()(global.(modelinst.ModelInst2), local); err != nil {
				return err
			}
			// process flow
			{
				for _, fn := range f.fnList {
					if err := fn(local.(pluginapi.Model)); err != nil {
						return err
					}
				}
			}
			dummy := modelinst.ModelInstHelper{}.NewInst()
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
				val := m.GetFieldUnsafe0(paths)
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
				val := m.GetFieldUnsafe0(paths)
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
				val := m.GetFieldUnsafe0(paths)
				val2 := m.GetFieldUnsafe0(paths2)
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
				val := m.GetFieldUnsafe0(paths)
				val2 := m.GetFieldUnsafe0(paths2)
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
				val := m.GetFieldUnsafe0(paths)
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
				val := m.GetFieldUnsafe0(paths)
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
