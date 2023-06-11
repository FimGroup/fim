package providers

type ContainerProvided interface {
	GetContainerLoggerManager() LoggerManager
}
