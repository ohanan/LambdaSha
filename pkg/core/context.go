package core

import (
	"iter"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/ohanan/LambdaSha/pkg/core/common"
	"github.com/ohanan/LambdaSha/pkg/lsha"
)

var _ lsha.RuntimeContext = (*runtimeContext)(nil)

func newContext(modeBuilder *modeBuilder, configData any, users []lsha.User) *Context {
	copiedUsers := make([]lsha.User, len(users))
	for i, user := range users {
		copiedUsers[i] = user
	}
	c := &Context{
		runtimeContext: &runtimeContext{
			modeBuilder:    modeBuilder,
			data:           atomic.Pointer[any]{},
			roomConfigData: configData,
			accounts:       copiedUsers,
		},
		parent: nil,
	}
	c.turn.Store(&Turn{})
	c.players.Store(common.Ptr([]*Player{}))
	return c
}

type runtimeContext struct {
	data               atomic.Pointer[any]
	players            atomic.Pointer[[]*Player]
	roomConfigData     any
	runtimeConfig      lsha.ConfigBuilder
	accounts           []lsha.User
	turn               atomic.Pointer[Turn]
	triggers           sync.Map // id -> *Trigger
	triggerByEventName sync.Map // eventName -> *sync.Map[uint64, *Trigger]
	triggerNextID      uint64
	triggerMutex       sync.Mutex
	modeBuilder        *modeBuilder
}

func (c *runtimeContext) AddTrigger(trigger lsha.Trigger, player lsha.Player, eventNames ...string) (id uint64) {
	if len(eventNames) == 0 || trigger == nil || trigger.Name() == "" {
		return 0
	}
	id = atomic.AddUint64(&c.triggerNextID, 1)
	t := &Trigger{
		id:           id,
		Trigger:      trigger,
		player:       player,
		eventNameMap: common.SliceToStructMap(eventNames),
	}
	c.triggers.Store(id, t)
	for _, eventName := range eventNames {
		m, ok := c.triggerByEventName.Load(eventName)
		if !ok {
			m, _ = c.triggerByEventName.LoadOrStore(eventName, &sync.Map{})
		}
		m.(*sync.Map).Store(id, t)
	}
	return c.triggerNextID
}

func (c *runtimeContext) NextPlayer(player lsha.Player) lsha.Player {
	players := *c.players.Load()
	if len(players) == 0 {
		return nil
	}
	if player == nil {
		player = c.turn.Load().player
	}
	if player == nil {
		player = players[0]
	}
	order := player.Order()
	for i := order + 1; i != order; i = (order + 1) % len(players) {
		p := players[i]
		if p.IsAlive() {
			return p
		}
	}
	return nil
}
func (c *runtimeContext) RemoveTrigger(id uint64) {
	value, loaded := c.triggers.LoadAndDelete(id)
	if !loaded {
		return
	}
	t := value.(*Trigger)
	for s := range t.eventNameMap {
		h, ok := c.triggerByEventName.Load(s)
		if !ok {
			return
		}
		h.(*sync.Map).Delete(id)
	}
}

func (c *runtimeContext) BindData(data any) {
	c.data.Store(&data)
}

func (c *runtimeContext) Data() any {
	return c.data.Load()
}
func (c *runtimeContext) RoomConfig() any { return c.roomConfigData }
func (c *runtimeContext) RuntimeConfig() lsha.ConfigBuilder {
	return c.runtimeConfig
}

func (c *runtimeContext) Turn() lsha.Turn {
	return c.turn.Load()
}

func (c *runtimeContext) PlayerIter(start lsha.Player) iter.Seq[lsha.Player] {
	return func(yield func(lsha.Player) bool) {
		players := *c.players.Load()
		var startIdx int
		for i, player := range players {
			if player.user.ID() == start.User().ID() {
				startIdx = i
				break
			}
		}
		player := players[startIdx]
		if player.IsAlive() && !yield(player) {
			return
		}
		size := len(players)
		for i := startIdx + 1; i != startIdx; i = (startIdx + 1) % size {
			player = players[i]
			if player.IsAlive() && !yield(player) {
				return
			}
		}
	}
}

type Context struct {
	*runtimeContext
	parent *Context
	event  lsha.Event
}

func (c *Context) Invoke(event lsha.Event) {
	var triggers []*Trigger
	if raw, ok := c.triggerByEventName.Load(event.Name()); ok {
		raw.(*sync.Map).Range(func(key, value any) bool {
			trigger := value.(*Trigger)
			if _, ok := c.triggers.Load(trigger.id); !ok {
				return true
			}
			if trigger.player == nil || trigger.player.IsAlive() {
				triggers = append(triggers, trigger)
			}
			return true
		})
	}
	startOrder := -1
	if wp, ok := event.(lsha.EventWithStartPlayer); ok {
		if sp := wp.StartPlayer(); sp != nil {
			startOrder = sp.Order()
		}
	}
	if startOrder < 0 {
		if p := c.turn.Load().player; p != nil {
			startOrder = p.order
		}
	}
	sort.Slice(triggers, func(i, j int) bool {
		t1, t2 := triggers[i], triggers[j]
		priority1, priority2 := t1.Priority(), t2.Priority()
		if priority1 < priority2 {
			return true
		}
		if priority1 > priority2 {
			return false
		}
		order1, order2 := t1.getTriggerPlayerOrder(), t2.getTriggerPlayerOrder()
		if order1 < 0 && order2 < 0 { // all system triggers
			return t1.id < t2.id
		}
		if order1 < 0 || order2 < 0 { // exist one system trigger
			return order1 < 0
		}
		if order1 == order2 { // same player trigger
			return t1.id < t2.id
		}
		if order1 == startOrder {
			return true
		}
		if order2 == startOrder {
			return false
		}
		// start order is in order1 and order2
		if order1 < startOrder && startOrder < order2 || order1 > startOrder && startOrder > order2 {
			return order1 > order2
		}
		return order1 < order2
	})
	ctx := c.WithEvent(event)
	for _, trigger := range triggers {
		r := &invokerResult{}
		trigger.Invoke(ctx, true, r)
	}
	for i := len(triggers) - 1; i >= 0; i-- {
		trigger := triggers[i]
		r := &invokerResult{}
		trigger.Invoke(ctx, false, r)
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
