package definition

type DataFlowDirection int

type internalDirectionIndicator int

const (
	DataFlowInOnly DataFlowDirection = iota + 1
	DataFlowInOut
)

type Process interface {
	DataFlowDirection() DataFlowDirection
}

type InOnlyProcess interface {
	typeIndicator() internalDirectionIndicator
}

type InOutProcess interface {
	typeIndicator() internalDirectionIndicator
}
