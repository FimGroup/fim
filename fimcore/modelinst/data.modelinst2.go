package modelinst

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/FimGroup/fim/fimapi/pluginapi"
	"github.com/FimGroup/fim/fimapi/rule"
)

type MappingRuleRaw [][]interface{}

// ToConverter generate ModelConverter
// Responsibility:
// 1. build path structure
// 2. do basic checks on path structure(primitive data type check excluded)
func (m MappingRuleRaw) ToConverter() (*ModelConverter, error) {

	converter := new(ModelConverter)

	for _, sub := range m {
		var ruleElem = sub

		if lp, err := m.processEachLevelOfRule(ruleElem, converter, nil, nil); err != nil {
			return nil, err
		} else {
			converter.LevelPair = append(converter.LevelPair, lp)
		}
	}

	//FIXME need to calculate max level cnt to optimize assignment
	converter.MaxLevelCnt = 16

	return converter, nil
}

func (m MappingRuleRaw) processEachLevelOfRule(ruleElem []interface{}, converter *ModelConverter, srcPaths, dstPaths []string) (levelPair, error) {
	switch len(ruleElem) {
	case 2:
		// direct assignment
		src, ok := ruleElem[0].(string)
		if !ok {
			return levelPair{}, errors.New("rule element is not string for source path")
		}
		dst, ok := ruleElem[1].(string)
		if !ok {
			return levelPair{}, errors.New("rule element is not string for destination path")
		}
		// check path empty
		if len(src) == 0 && len(dst) == 0 {
			return levelPair{}, errors.New("both path should be not empty at the same time")
		}
		// check path format
		if len(src) > 0 && !rule.ValidateFullPathOfDefinition(src) {
			return levelPair{}, errors.New("invalid path:" + src)
		}
		if len(dst) > 0 && !rule.ValidateFullPathOfDefinition(dst) {
			return levelPair{}, errors.New("invalid path:" + dst)
		}
		// validations
		if rule.IsPathArray(src) {
			return levelPair{}, errors.New("src array is not allowed in direct assignment")
		}
		if rule.IsPathArray(dst) {
			return levelPair{}, errors.New("dst array is not allowed in direct assignment")
		}

		srcPath := rule.ConcatFullPath(append(srcPaths, src))
		dstPath := rule.ConcatFullPath(append(dstPaths, dst))
		converter.SourceLeafPathList = append(converter.SourceLeafPathList, srcPath)
		converter.TargetLeafPathList = append(converter.TargetLeafPathList, dstPath)

		// generate converter materials
		lp := levelPair{}
		{
			lp.Leaf = true
			{
				// only field is allowed
				lp.Src = src
				lp.SrcName = src
				lp.SrcArray = false
			}
			{
				// only field is allowed
				lp.Dst = dst
				lp.DstName = dst
				lp.DstArray = false
			}
		}

		return lp, nil
	case 3:
		// object/array assignment
		src, ok := ruleElem[0].(string)
		if !ok {
			return levelPair{}, errors.New("rule element is not string for source path")
		}
		dst, ok := ruleElem[1].(string)
		if !ok {
			return levelPair{}, errors.New("rule element is not string for destination path")
		}
		// check path empty
		if len(src) == 0 && len(dst) == 0 {
			return levelPair{}, errors.New("both path should be not empty at the same time")
		}
		subs, ok := ruleElem[2].([]interface{})
		if !ok {
			return levelPair{}, errors.New("rule element is not sub rule for third")
		}
		// check path format
		if len(src) > 0 && !rule.ValidateFullPathOfDefinition(src) {
			return levelPair{}, errors.New("invalid path:" + src)
		}
		if len(dst) > 0 && !rule.ValidateFullPathOfDefinition(dst) {
			return levelPair{}, errors.New("invalid path:" + dst)
		}
		// validations
		// nothing

		// skip empty path
		newSrcPaths := srcPaths
		if len(src) > 0 {
			newSrcPaths = append(newSrcPaths, src)
		}
		newDstPaths := dstPaths
		if len(dst) > 0 {
			newDstPaths = append(newDstPaths, dst)
		}
		// process sub level
		lp := levelPair{}
		for _, v := range subs {
			newRuleElem, ok := v.([]interface{})
			if !ok {
				return levelPair{}, errors.New("sub rule is not rule definition spec")
			}
			if subLp, err := m.processEachLevelOfRule(newRuleElem, converter, newSrcPaths, newDstPaths); err != nil {
				return levelPair{}, err
			} else {
				lp.Subs = append(lp.Subs, subLp)
			}
		}

		// generate converter materials
		{
			lp.Leaf = false
			{
				lp.Src = src
				name, _ := rule.ExtractArrayPath(src)
				lp.SrcName = name
				if rule.IsPathArray(src) {
					lp.SrcArray = true
				} else {
					// type
					lp.SrcArray = false
				}
			}
			{
				lp.Dst = dst
				name, _ := rule.ExtractArrayPath(dst)
				lp.DstName = name
				if rule.IsPathArray(dst) {
					lp.DstArray = true
				} else {
					// type
					lp.DstArray = false
				}
			}
		}
		// special case - primitive array(array to array) with empty mapping rules
		if rule.IsArrayDefinition(src) && rule.IsArrayDefinition(dst) && len(subs) == 0 {
			lp.Leaf = true
		}
		// not allow only one side is array
		if lp.SrcArray != lp.DstArray {
			return levelPair{}, errors.New("only one side is array")
		}

		return lp, nil
	default:
		return levelPair{}, errors.New("rule element size is not 2/3 which means not direct or object/array assignment")
	}
}

type ModelConverter struct {
	SourceLeafPathList []string
	TargetLeafPathList []string

	LevelPair   []levelPair
	MaxLevelCnt int
}

// ModelInst2 contains full definition of data object
// Note: when moving data(transfer/assign), data type should be performed according tothe certain operation
type ModelInst2 interface {
	transferPrimitiveArray(srcName, dstName string, dst ModelInst2) error
	transferValue(srcName, dstName string, dst ModelInst2) error
	ensureSubObject(name string) (ModelInst2, error)
	getSubObject(name string) (ModelInst2, error)
	ensureSubArrayWithObjectElem(name string) (ModelInst2, error)
	getSubArrayWithObjectElem(name string) (ModelInst2, error)

	foreachArrayElement(func(inst2 ModelInst2) error) error
	ensureArrayElement() (ModelInst2, error)
	//ensureArrayElementWithIndex will ensure the item of the given index responded
	// if the array is not large enough, empty element will be added until the give index element can be retrieved
	ensureArrayElementWithIndex(idx int) (ModelInst2, error)

	putPrimitiveArray(name string, arr []interface{}) error
	setPrimitiveArrayIndex(name string, index int, value interface{}) error
	putPrimitiveValue(name string, val interface{}) error

	RemoveObjectByPath(paths []string) error

	pluginapi.Model
}

type ModelInstHelper struct {
}

func (ModelInstHelper) WrapMap(m map[string]interface{}) ModelInst2 {
	return mapWrapper{m: m}
}

func (ModelInstHelper) NewInst() ModelInst2 {
	return &modelInst2MapImpl{
		data:      map[string]*modelInst2MapImpl{},
		valueType: valueTypeObject,
	}
}

// Transfer do data Transfer according to converter definition
// responsibility:
// 1. transfer data
// 2. do data checks including primitive data type matching
func (m *ModelConverter) Transfer(src, dst ModelInst2) error {
	// DFS assignment + data type matching
	// rules of a given name
	// 1. src/dst literal types - field/object name, array def(xxx[]) name
	// 3. type array def -> object, primitives
	// 5. one of the two sides is empty - create level but no data assign

	for _, v := range m.LevelPair {
		if err := m.doTransfer(v, src, dst); err != nil {
			return err
		}
	}

	return nil
}

func (m *ModelConverter) doTransfer(lvPair levelPair, srcParent, dstParent ModelInst2) error {
	// leaf, do value assignment
	if lvPair.Leaf {
		// primitive array
		if lvPair.isPrimitiveArray() {
			return srcParent.transferPrimitiveArray(lvPair.SrcName, lvPair.DstName, dstParent)
		} else {
			// Transfer value
			return srcParent.transferValue(lvPair.SrcName, lvPair.DstName, dstParent)
		}
	}

	// if one side of each is empty, do recursive Transfer
	// otherwise do level preparation
	if len(lvPair.Src) == 0 {
		// only object allowed, array is not allowed
		if newDst, err := dstParent.ensureSubObject(lvPair.DstName); err != nil {
			return err
		} else {
			for _, sub := range lvPair.Subs {
				if err := m.doTransfer(sub, srcParent, newDst); err != nil {
					return err
				}
			}
			return nil
		}
	} else if len(lvPair.Dst) == 0 {
		// only object allowed, array is not allowed
		if subSrc, err := srcParent.getSubObject(lvPair.SrcName); err != nil {
			return err
		} else if subSrc == nil {
			// no src sub, just break deeper mapping
			return nil
		} else {
			for _, sub := range lvPair.Subs {
				if err := m.doTransfer(sub, subSrc, dstParent); err != nil {
					return err
				}
			}
			return nil
		}
	} else {
		// not leaf, just create levels
		if lvPair.SrcArray && lvPair.DstArray {
			// array to array
			srcArr, err := srcParent.getSubArrayWithObjectElem(lvPair.SrcName)
			if err != nil {
				return err
			} else if srcArr == nil {
				return nil
			}
			dstArr, err := dstParent.ensureSubArrayWithObjectElem(lvPair.DstName)
			if err != nil {
				return err
			}
			// foreach items and trigger sub mappings, then fill items back in the array
			if err := srcArr.foreachArrayElement(func(srcObj ModelInst2) error {
				if dstObj, err := dstArr.ensureArrayElement(); err != nil {
					return err
				} else {
					for _, sub := range lvPair.Subs {
						if err := m.doTransfer(sub, srcObj, dstObj); err != nil {
							return err
						}
					}
					return nil
				}
			}); err != nil {
				return err
			}
			return nil
		} else {
			// object to object
			srcObj, err := srcParent.getSubObject(lvPair.SrcName)
			if err != nil {
				return err
			} else if srcObj == nil {
				return nil
			}
			dstObj, err := dstParent.ensureSubObject(lvPair.DstName)
			if err != nil {
				return err
			}
			for _, sub := range lvPair.Subs {
				if err := m.doTransfer(sub, srcObj, dstObj); err != nil {
					return err
				}
			}
			return nil
		} // otherwise not allowed
	}

	// For this very complex mapping rule, in order to avoid missing branches, 'return nil' will not be used at the bottom
	// Instead, 'return' will be used in every branchy
}

type levelPair struct {
	Leaf bool
	Subs []levelPair

	Src      string
	SrcName  string
	SrcArray bool
	Dst      string
	DstName  string
	DstArray bool
}

func (l *levelPair) isPrimitiveArray() bool {
	return l.Leaf && l.SrcArray && l.DstArray
}

func isPrimitive(in interface{}) bool {
	switch in.(type) {
	case float64:
	case bool:
	case string:
	case int64:
	default:
		return false
	}
	return true
}

func defaultValuePrimitive(in interface{}) interface{} {
	switch in.(type) {
	case float64:
		return 0.0
	case bool:
		return false
	case string:
		return ""
	case int64:
		return 0
	default:
		panic(errors.New("unexpected primitive type:" + fmt.Sprint(reflect.TypeOf(in))))
	}
}

func convertPrimitive(in interface{}) (interface{}, error) {
	switch v := in.(type) {
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	case bool:
		return v, nil
	case uint8:
		return int64(v), nil
	case uint16:
		return int64(v), nil
	case uint32:
		return int64(v), nil
	case uint64:
		return nil, errors.New("currently uint64 is not supported for conversion")
	case int8:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	case complex64:
		return nil, errors.New("currently complex64 is not supported for conversion")
	case complex128:
		return nil, errors.New("currently complex128 is not supported for conversion")
	case string:
		return v, nil
	case int:
		return int64(v), nil
	case uint:
		return nil, errors.New("currently uint is not supported for conversion")
	case uintptr:
		return nil, errors.New("currently uintptr is not supported for conversion")
	}
	return nil, errors.New("unknown type for conversion:" + fmt.Sprint(reflect.TypeOf(in)))
}

func mustConvertPrimitive(in interface{}) interface{} {
	v, err := convertPrimitive(in)
	if err != nil {
		panic(err)
	}
	return v
}
