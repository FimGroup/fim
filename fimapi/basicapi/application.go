package basicapi

type Application interface {
	SpawnUseContainer(businessName string) BasicContainer
}
