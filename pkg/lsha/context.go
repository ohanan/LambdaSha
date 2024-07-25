package lsha

import "iter"

type Context interface {
	RuntimeContext
	WithEvent(event Event) Context
	Event() Event
}
type RuntimeContext interface {
	BindData(data any)
	Data() any
	RoomConfig() any
	RuntimeConfig() ConfigBuilder
	Turn() Turn
	PlayerIter(start Player) iter.Seq[Player]
	NextPlayer(start Player) Player
	AddTrigger(trigger Trigger, player Player, eventNames ...string) (id uint64)
	RemoveTrigger(id uint64)
}
type DataHolder interface {
	BindData(data any)
	Data() any
}

func Data[V any](c DataHolder) (_ V) {
	if v := c.Data(); v != nil {
		if v, ok := v.(V); ok {
			return v
		}
	}
	return
}
func TurnData[V any](ctx Context) (_ V) {
	if v := ctx.Turn(); v != nil {
		if v, ok := v.(V); ok {
			return v
		}
	}
	return
}

func PhaseData[V any](ctx Context) (_ V) {
	if t := ctx.Turn(); t != nil {
		if v := t.Phase(); v != nil {
			if v, ok := v.(V); ok {
				return v
			}
		}
	}
	return
}
