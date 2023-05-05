package esbcore

type TypeOfNode int

const (
	TypeUnknown       TypeOfNode = 0
	TypeDataNode      TypeOfNode = 1
	TypeNsNode        TypeOfNode = 2
	TypeAttributeNode TypeOfNode = 3
)

type ModelInst struct {
}

func (m *ModelInst) AddField(path string, value interface{}) {
	panic(_IMPLEMENT_ME)
}

func (m *ModelInst) UpdateField(path string, value interface{}) {
	panic(_IMPLEMENT_ME)
}

func (m *ModelInst) DeleteField(path string) {
	panic(_IMPLEMENT_ME)
}

func (m *ModelInst) GetFields() []string {
	panic(_IMPLEMENT_ME)
}

func (m *ModelInst) GetField(path string) interface{} {
	panic(_IMPLEMENT_ME)
}
