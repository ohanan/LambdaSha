package core

import (
	"container/heap"
	"sync"
	"sync/atomic"

	"github.com/ohanan/LambdaSha/pkg/core/form"
	"github.com/ohanan/LambdaSha/pkg/lsha"
)

var _ lsha.RuntimeContext = (*runtimeContext)(nil)

func newContext(roomConfig *form.ItemsBuilder, users *[]*User) *Context {
	copiedUsers := make([]lsha.User, len(*users))
	for i, user := range *users {
		copiedUsers[i] = user
	}
	c := &Context{
		runtimeContext: &runtimeContext{
			data:          atomic.Pointer[any]{},
			roomConfig:    roomConfig,
			runtimeConfig: nil,
			accounts:      copiedUsers,
			turn:          atomic.Pointer[Turn]{},
		},
		parent: nil,
	}
	c.turn.Store(&Turn{})
	return c
}

type runtimeContext struct {
	data               atomic.Pointer[any]
	roomConfig         lsha.ConfigBuilder
	runtimeConfig      lsha.ConfigBuilder
	accounts           []lsha.User
	turn               atomic.Pointer[Turn]
	triggers           map[uint64]*triggerWithID
	triggerByEventName map[string]*triggerHeap
	triggerNextID      uint64
	triggerMutex       sync.Mutex
}
type triggerWithID struct {
	t            lsha.Trigger
	id           uint64
	eventNameMap map[string]struct{}
}

func (c *runtimeContext) AddTrigger(t lsha.Trigger, eventNames ...string) (id uint64) {
	if len(eventNames) == 0 {
		return 0
	}
	c.triggerMutex.Lock()
	defer c.triggerMutex.Unlock()
	c.triggerNextID++
	id = c.triggerNextID
	innerTrigger := &triggerWithID{t: t, id: id, eventNameMap: map[string]struct{}{}}
	c.triggers[id] = innerTrigger
	for _, eventName := range eventNames {
		h, ok := c.triggerByEventName[eventName]
		if !ok {
			h = &triggerHeap{}
			c.triggerByEventName[eventName] = h
		}
		heap.Push(h, t)
	}
	return c.triggerNextID
}

func (c *runtimeContext) RemoveTrigger(id uint64) {
	c.triggerMutex.Lock()
	defer c.triggerMutex.Unlock()
	t, ok := c.triggers[id]
	if !ok {
		return
	}
	delete(c.triggers, id)
	for s := range t.eventNameMap {
		h, ok := c.triggerByEventName[s]
		if !ok {
			return
		}
		for i, trigger := range *h {
			if trigger.id == id {
				heap.Remove(h, i)
			}
		}
	}
}

func (c *runtimeContext) BindData(data any) {
	c.data.Store(&data)
}

func (c *runtimeContext) Data() any {
	return c.data.Load()
}

func (c *runtimeContext) RoomConfig() lsha.ConfigBuilder {
	return c.roomConfig
}

func (c *runtimeContext) RuntimeConfig() lsha.ConfigBuilder {
	return c.runtimeConfig
}

func (c *runtimeContext) Users() []lsha.User {
	return c.accounts
}

func (c *runtimeContext) Turn() lsha.Turn {
	return c.turn.Load()
}

type Context struct {
	*runtimeContext
	parent *Context
	event  lsha.Event
}

func (c *Context) Invoke(event lsha.Event) {
	if triggers, ok := c.triggerByEventName[event.Name()]; ok {
		ctx := c.WithEvent(event)
		for _, trigger := range *triggers {
			result := &InvokerResult{}
			trigger.t.Invoke(ctx, result)
		}
	}

}

func (c *Context) WithEvent(event lsha.Event) lsha.Context {
	return &Context{
		runtimeContext: c.runtimeContext,
		parent:         c,
		event:          event,
	}
}

func (c *Context) Event() lsha.Event {
	return c.event
}
