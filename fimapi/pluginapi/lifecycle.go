package pluginapi

type LifecycleListener interface {
	OnStart() error
	OnStop() error
}
