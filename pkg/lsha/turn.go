package lsha

type (
	Turn          = any
	FuncTurnStart = func(ctx Context) Turn
)

type TurnBuilder interface {
	Player(p Player)
	Round(n int)
	Start(f FuncModeStart)
}
