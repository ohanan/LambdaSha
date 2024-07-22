package lsha

type Context interface {
	RuntimeContext
	WithEvent(event Event) Context
	Event() Event
}
type RuntimeContext interface {
	BindData(data any)
	Data() any
	Users() []User
	RoomConfig() ConfigBuilder
	RuntimeConfig() ConfigBuilder
	Turn() Turn
	AddTrigger(trigger Trigger, eventNames ...string) (id uint64)
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
