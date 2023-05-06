package esbcore

import (
	"bytes"
	"errors"

	"github.com/pelletier/go-toml/v2"
)

type templateFlow struct {
	In     map[string]string                     `toml:"in"`
	Out    map[string]string                     `toml:"out"`
	PreOut map[string]string                     `toml:"pre_out"`
	Flow   map[string][]map[string][]interface{} `toml:"flow"`
}

type Flow struct {
	dtd *DataTypeDefinitions

	localInMapping map[string]struct {
		ModelFieldPath string
		DataType       DataType
	}
	localOutMapping map[string]struct {
		ModelFieldPath string
		DataType       DataType
	}
	localPreOutOperations map[string]struct {
		ModelFieldPath string
	}

	fnList []Fn
}

func NewFlow(dtd *DataTypeDefinitions) *Flow {
	return &Flow{
		dtd: dtd,

		localInMapping: map[string]struct {
			ModelFieldPath string
			DataType       DataType
		}{},
		localOutMapping: map[string]struct {
			ModelFieldPath string
			DataType       DataType
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
		if err := f.AddIn(path, local); err != nil {
			return err
		}
	}
	for local, path := range tf.Out {
		if err := f.AddOut(local, path); err != nil {
			return err
		}
	}
	if err := f.validateLocalParameters(); err != nil {
		return err
	}
	if err := f.AddFlow(tf); err != nil {
		return err
	}

	return nil
}

func (f *Flow) AddIn(source, local string) error {
	_, ok := f.localInMapping[local]
	if ok {
		return errors.New("local parameter registered:" + local)
	}

	if !ValidateFullPath(source) {
		return errors.New("in parameter path invalid:" + source)
	}

	if dt, _, err := f.dtd.TypeOfPath(source); err != nil {
		return err
	} else if dt == DataTypeUnavailable {
		return errors.New("cannot find path:" + source)
	} else {
		f.localInMapping[local] = struct {
			ModelFieldPath string
			DataType       DataType
		}{ModelFieldPath: source, DataType: dt}
	}

	return nil
}

func (f *Flow) inConv() func(source, local *ModelInst) error {
	return func(source, local *ModelInst) error {
		for s, dStruct := range f.localInMapping {
			if err := source.transferTo(local, s, dStruct.ModelFieldPath, ByRight); err != nil {
				return err
			}
		}
		return nil
	}
}

func (f *Flow) AddOut(local, out string) error {
	_, ok := f.localOutMapping[local]
	if ok {
		return errors.New("local parameter registered:" + local)
	}

	if !ValidateFullPath(out) {
		return errors.New("out parameter path invalid:" + out)
	}

	if dt, _, err := f.dtd.TypeOfPath(out); err != nil {
		return err
	} else if dt == DataTypeUnavailable {
		return errors.New("cannot find path:" + out)
	} else {
		f.localOutMapping[local] = struct {
			ModelFieldPath string
			DataType       DataType
		}{ModelFieldPath: out, DataType: dt}
	}

	return nil
}

func (f *Flow) outConv() func(local, out *ModelInst) error {
	return func(local, out *ModelInst) error {
		// process pre_out
		for op, s := range f.localPreOutOperations {
			switch op {
			case "@remove":
				if err := out.deleteField(s.ModelFieldPath); err != nil {
					return err
				}
			default:
				return errors.New("unknown pre_out operation:" + op)
			}
		}

		// process out
		for s, dStruct := range f.localOutMapping {
			if err := local.transferTo(out, s, dStruct.ModelFieldPath, ByRight); err != nil {
				return err
			}
		}

		return nil
	}
}

func (f *Flow) FlowFn() func(global *ModelInst) error {
	local := NewModelInst(f.dtd)
	return func(global *ModelInst) error {
		if err := f.inConv()(global, local); err != nil {
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
		if err := f.outConv()(local, global); err != nil {
			return err
		}

		return nil
	}
}

func (f *Flow) AddFlow(tf *templateFlow) error {
	steps := tf.Flow["steps"]
	var fList []Fn
	for _, step := range steps {
		for fn, params := range step {
			if fn[0] == '@' {
				//builtin function
				fngen, ok := builtinGenFnMap[fn]
				if !ok {
					return errors.New("function not found:" + fn)
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
