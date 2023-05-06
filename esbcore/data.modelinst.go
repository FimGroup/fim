package esbcore

type TypeOfNode int
type ElementMap map[string]interface{}

const (
	TypeUnknown       TypeOfNode = 0
	TypeDataNode      TypeOfNode = 1
	TypeNsNode        TypeOfNode = 2
	TypeAttributeNode TypeOfNode = 3
)

type ModelInst struct {
	dtd        *DataTypeDefinitions
	ElementMap ElementMap
}

func NewModelInst(def *DataTypeDefinitions) *ModelInst {
	return &ModelInst{
		dtd: def,
	}
}

func (m *ModelInst) addOrUpdateField(path string, value interface{}) error {
	panic(_IMPLEMENT_ME)
}

func (m *ModelInst) deleteField(path string) error {
	panic(_IMPLEMENT_ME)
}

func (m *ModelInst) getField(path string) interface{} {
	splits := SplitFullPath(path)

	var result interface{} = m.ElementMap
	for _, pLv := range splits {
		pathName, idx := ExtractArrayPath(pLv)
		isArrAccess := idx >= 0
		elemMap, ok := result.(ElementMap)
		if !ok {
			return nil
		}
		elem, ok := elemMap[pathName]
		if !ok {
			return nil
		}
		if isArrAccess {
			result = elem.([]ElementMap)[idx]
		} else {
			result = elem
		}
	}

	return result
}

func (m *ModelInst) FillIn(o interface{}) error {
	panic(_IMPLEMENT_ME)
}

func (m *ModelInst) transferTo(dest *ModelInst, sourcePath, destPath string) error {
	val := m.getField(sourcePath)
	if val != nil {
		return dest.addOrUpdateField(destPath, val)
	}
	return nil
}
