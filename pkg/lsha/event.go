package lsha

const (
	EventPlayerPrepared = "system:player_prepared"
	EventGameStarted    = "system:game_start"
	EventTurnStarted    = "system:turn_start"
	EventPhaseStarted   = "system:phase_start"
)

type (
	Invoker = func(ctx Context, result InvokeResult)
)
type Event interface {
	Name() string
}
type EventWithStartPlayer interface {
	Event
	StartPlayer() Player
}
type Trigger interface {
	Name() string
	EventName() string
	Priority() float64
	Invoke(ctx Context, enter bool, result InvokeResult)
}
type InvokeResult interface {
	FastStop()
}
type GameStartedEvent struct {
}

type PlayerPreparedEvent struct {
	player Player
}

func (e *PlayerPreparedEvent) Player() Player          { return e.player }
func (e *PlayerPreparedEvent) SetPlayer(player Player) { e.player = player }

type TurnStartedEvent struct {
	turn Turn
}

func (e *TurnStartedEvent) Turn() Turn        { return e.turn }
func (e *TurnStartedEvent) SetTurn(turn Turn) { e.turn = turn }

type PhaseStartedEvent struct {
	phase Phase
	turn  Turn
}

func (e *PhaseStartedEvent) Turn() Turn           { return e.turn }
func (e *PhaseStartedEvent) SetTurn(turn Turn)    { e.turn = turn }
func (e *PhaseStartedEvent) Phase() Phase         { return e.phase }
func (e *PhaseStartedEvent) SetPhase(phase Phase) { e.phase = phase }

func (e *GameStartedEvent) Name() string    { return EventGameStarted }
func (e *PlayerPreparedEvent) Name() string { return EventPlayerPrepared }
func (e *TurnStartedEvent) Name() string    { return EventTurnStarted }
func (e *PhaseStartedEvent) Name() string   { return EventPhaseStarted }

func (e *PlayerPreparedEvent) StartPlayer() Player { return e.player }
func (e *TurnStartedEvent) StartPlayer() Player    { return e.turn.Player() }
func (e *PhaseStartedEvent) StartPlayer() Player   { return e.turn.Player() }
