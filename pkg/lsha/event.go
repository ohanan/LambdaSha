package lsha

const (
	EventNameGameStarting  = "system:game_starting"
	EventNameGameStarted   = "system:game_started"
	EventNameHeroAppear    = "system:hero_appear"
	EventNameTurnStart     = "system:turn_start"
	EventNameTurnPreCheck  = "system:turn_pre_check"
	EventNameTurnHarvest   = "system:turn_harvest"
	EventNameTurnPlay      = "system:turn_play"
	EventNameTurnPostCheck = "system:turn_post_check"
	EventNameTurnEnd       = "system:turn_end"
)

type (
	Invoker = func(ctx Context, result InvokeResult)
)
type Event interface {
	Name() string
}
type Trigger interface {
	Name() string
	EventName() string
	Priority() float64
	Invoke(ctx Context, result InvokeResult)
}
type InvokeResult interface {
	FastStop()
}

type GameStartingEvent struct {
}

type GameStartedEvent struct {
}

type AppearEvent struct {
	Hero Hero
}

type TurnStartEvent struct{}
type TurnPreCheckEvent struct{}
type TurnHarvestEvent struct{}
type TurnPlayEvent struct{}
type TurnPostCheckEvent struct{}
type TurnEndEvent struct {
}

func (e *GameStartingEvent) Name() string  { return EventNameGameStarting }
func (e *GameStartedEvent) Name() string   { return EventNameGameStarted }
func (e *AppearEvent) Name() string        { return EventNameHeroAppear }
func (e *TurnStartEvent) Name() string     { return EventNameTurnStart }
func (e *TurnPreCheckEvent) Name() string  { return EventNameTurnPreCheck }
func (e *TurnHarvestEvent) Name() string   { return EventNameTurnHarvest }
func (e *TurnPlayEvent) Name() string      { return EventNameTurnPlay }
func (e *TurnPostCheckEvent) Name() string { return EventNameTurnPostCheck }
func (e *TurnEndEvent) Name() string       { return EventNameTurnEnd }
