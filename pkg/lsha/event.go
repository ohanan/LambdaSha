package lsha

import (
	"sync"
	"sync/atomic"
)

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

type Event interface {
	Name() string
}
type EventManager interface {
	AddTrigger(trigger Trigger) (id uint64)
	RemoveTrigger(id uint64)
	GetTrigger(id uint64)
	Invoke(event Event)
}
type Trigger interface {
	TriggerMeta
	RawInvoke(event Event, result TriggerResult)
}
type TriggerMeta interface {
	ID() uint64
	Name() string
	EventName() string
	Priority() float64
}
type TriggerResult interface {
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

type triggerManager struct {
	nextID   uint64
	triggers sync.Map
}

func (m *triggerManager) AddTrigger(trigger Trigger) (id uint64) {
	if trigger == nil {
		return
	}
	id = atomic.AddUint64(&m.nextID, 1)
	m.triggers.Store(id, trigger)
	return id
}

func (m *triggerManager) RemoveTrigger(id uint64) {
	m.triggers.Delete(id)
}

func (m *triggerManager) GetTrigger(id uint64) (Trigger, bool) {
	if id == 0 {
		return nil, false
	}
	v, ok := m.triggers.Load(id)
	if !ok {
		return nil, false
	}
	return v.(Trigger), true
}

func (m *triggerManager) Invoke(event Event) {
	if event == nil {
		return
	}
	m.triggers.Range(func(key, value any) bool {
		trigger := value.(Trigger)
		if trigger.EventName() == event.Name() {
			trigger.RawInvoke(event, nil)
		}
		return true
	})
}
