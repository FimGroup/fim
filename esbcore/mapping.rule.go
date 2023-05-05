package esbcore

type ConversionMappingInst struct {
}

func (c *ConversionMappingInst) AddMapping(pathFrom, pathTo string) error {
	panic(_IMPLEMENT_ME)
}

func (c *ConversionMappingInst) FindMappingByFrom(pathFrom string) string {
	panic(_IMPLEMENT_ME)
}
