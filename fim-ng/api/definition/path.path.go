package definition

type PathSettings struct {
}

type PathBuilder interface {
	Configure()
}

type PathDef struct {
}

func (p *PathDef) NewPath(s PathSettings) *PathInstance {
	//TODO
	return &PathInstance{}
}

func (p *PathDef) Configure() {
	panic("please implement PathDef.Configure()")
}

type PathInstance struct {
}

func (i *PathInstance) Next() *PathInstance {
	//TODO
	return i
}
