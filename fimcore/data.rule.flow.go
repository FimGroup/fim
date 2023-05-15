package fimcore

import (
	"bytes"
	"errors"

	"github.com/ThisIsSun/fim/fimapi/pluginapi"
	"github.com/ThisIsSun/fim/fimapi/rule"

	"github.com/pelletier/go-toml/v2"
)

type templateFlow struct {
	In     map[string]string                     `toml:"in"`
	Out    map[string]string                     `toml:"out"`
	PreOut map[string]string                     `toml:"pre_out"`
	Flow   map[string][]map[string][]interface{} `toml:"flow"`
}

type Flow struct {
	dtd       *DataTypeDefinitions
	container *ContainerInst

	localInMapping map[string]struct {
		ModelFieldPath string
		SplitPath      []string
		DataType       pluginapi.DataType
		KeySplitPath   []string
	}
	localOutMapping map[string]struct {
		ModelFieldPath string
		SplitPath      []string
		DataType       pluginapi.DataType
		KeySplitPath   []string
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

		localInMapping: map[string]struct {
			ModelFieldPath string
			SplitPath      []string
			DataType       pluginapi.DataType
			KeySplitPath   []string
		}{},
		localOutMapping: map[string]struct {
			ModelFieldPath string
			SplitPath      []string
			DataType       pluginapi.DataType
			KeySplitPath   []string
		}{},
		localPreOutOperations: map[string]struct {
			Operation string
			SplitPath []string
		}{},
	}
}

func (f *Flow) MergeToml(data string) error {
	tf := new(templateFlow)
	err := toml.NewDecoder(bytes.NewBufferString(data)).DisallowUnknownFields().Decode(tf)
	if err != nil {
		return err
	}

	for path, local := range tf.In {
		if err := f.addIn(path, local); err != nil {
			return err
		}
	}
	for local, path := range tf.Out {
		if err := f.addOut(local, path); err != nil {
			return err
		}
	}
	for op, path := range tf.PreOut {
		if err := f.addPreOut(op, path); err != nil {
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
	_, ok := f.localInMapping[local]
	if ok {
		return errors.New("local parameter registered:" + local)
	}

	if !rule.ValidateFullPath(source) {
		return errors.New("in parameter path invalid:" + source)
	}

	if dt, _, err := f.dtd.TypeOfPath(source); err != nil {
		return err
	} else if dt == pluginapi.DataTypeUnavailable {
		return errors.New("cannot find path:" + source)
	} else {
		f.localInMapping[local] = struct {
			ModelFieldPath string
			SplitPath      []string
			DataType       pluginapi.DataType
			KeySplitPath   []string
		}{ModelFieldPath: source, SplitPath: rule.SplitFullPath(source), DataType: dt, KeySplitPath: rule.SplitFullPath(local)}
	}

	return nil
}

func (f *Flow) inConv() func(source, local *ModelInst) error {
	return func(source, local *ModelInst) error {
		for _, dStruct := range f.localInMapping {
			if err := source.transferTo(local, dStruct.SplitPath, dStruct.KeySplitPath, ByLeft); err != nil {
				return err
			}
		}
		return nil
	}
}

func (f *Flow) addOut(local, out string) error {
	_, ok := f.localOutMapping[local]
	if ok {
		return errors.New("local parameter registered:" + local)
	}

	if !rule.ValidateFullPath(out) {
		return errors.New("out parameter path invalid:" + out)
	}

	if dt, _, err := f.dtd.TypeOfPath(out); err != nil {
		return err
	} else if dt == pluginapi.DataTypeUnavailable {
		return errors.New("cannot find path:" + out)
	} else {
		f.localOutMapping[local] = struct {
			ModelFieldPath string
			SplitPath      []string
			DataType       pluginapi.DataType
			KeySplitPath   []string
		}{ModelFieldPath: out, SplitPath: rule.SplitFullPath(out), DataType: dt, KeySplitPath: rule.SplitFullPath(local)}
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
		for _, dStruct := range f.localOutMapping {
			if err := local.transferTo(out, dStruct.KeySplitPath, dStruct.SplitPath, ByRight); err != nil {
				return err
			}
		}

		return nil
	}
}

func (f *Flow) FlowFn() func(global pluginapi.Model) error {
	local := NewModelInst(f.dtd)
	return func(global pluginapi.Model) error {
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

func (f *Flow) FlowFnNoResp() func(global pluginapi.Model) error {
	local := NewModelInst(f.dtd)
	return func(global pluginapi.Model) error {
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

func (f *Flow) addFlow(tf *templateFlow) error {
	steps := tf.Flow["steps"]
	var fList []pluginapi.Fn
	for _, step := range steps {
		for fn, params := range step {
			if fn[0] == '@' {
				//builtin function
				fngen, ok := f.container.builtinGenFnMap[fn]
				if !ok {
					return errors.New("builtin function not found:" + fn)
				}
				fnInst, err := fngen(params)
				if err != nil {
					return err
				}
				fList = append(fList, fnInst)
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
				fList = append(fList, fnInst)
			} else {
				return errors.New("unknown command:" + fn)
			}
			break
		}
	}
	f.fnList = fList
	return nil
}

func (f *Flow) validateLocalParameters() error {
	// compare the types of the same local parameters
	for local, si := range f.localInMapping {
		do, ok := f.localOutMapping[local]
		if ok {
			sdt, spdt, err := f.dtd.TypeOfPath(si.ModelFieldPath)
			if err != nil {
				return err
			}
			ddt, dpdt, err := f.dtd.TypeOfPath(do.ModelFieldPath)
			if err != nil {
				return err
			}
			if sdt != ddt || spdt != dpdt {
				return errors.New("local parameter types of in and out do not match:" + local)
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
