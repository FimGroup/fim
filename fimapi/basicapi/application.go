package basicapi

type Application interface {
	SpawnUseContainer() BasicContainer
}
