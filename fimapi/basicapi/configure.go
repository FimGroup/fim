package basicapi

type ConfigureManager interface {
	ReplaceStaticConfigure(placeholder string) string
	ReplaceDynamicConfigure(placeholder string) string
	SupportDynamicConfigure(placeholder string) bool
}
