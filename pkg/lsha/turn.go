package lsha

type (
	TurnStarter = func(ctx Context, tb TurnBuilder) (turnData any)
)

type Turn interface {
	DataHolder
	Player() Player
	Round() int
	Phase() Phase
}

type TurnBuilder interface {
	Player(p Player) TurnBuilder
	Round(n int) TurnBuilder
	OnNextPhase(phaseStarter PhaseStarter) TurnBuilder
}
