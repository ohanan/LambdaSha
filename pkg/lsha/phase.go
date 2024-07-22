package lsha

type (
	PhaseStarter = func(ctx Context, pb PhaseBuilder) (phaseData any)
)
type Phase interface {
	DataHolder
	Name() string
}

type PhaseBuilder interface {
	Name(name string) PhaseBuilder
}
