package core

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/ohanan/LambdaSha/pkg/lsha"
)

func NewHandler() *Handler {
	return &Handler{
		modeBuilders:   map[string]*modeBuilder{},
		pluginBuilders: map[string]*pluginBuilder{},
	}
}

type Handler struct {
	modeBuilders   map[string]*modeBuilder
	pluginBuilders map[string]*pluginBuilder

	rooms      sync.Map
	roomNextID int64
}

func (h *Handler) Init(initializers []func(builder lsha.PluginBuilder)) {
	for _, initializer := range initializers {
		b := &pluginBuilder{}
		initializer(b)
		name := b.name
		if name == "" {
			h.Warning("plugin name is missing")
			continue
		}
		if _, ok := h.pluginBuilders[name]; ok {
			h.Warning("pluginBuilder %s already exists", name)
			continue
		}
		h.pluginBuilders[name] = b
	}
	h.loadPlugins()
}

func (h *Handler) GetModeRegistration(name string) lsha.ModeRegistration {
	return h.modeBuilders[name]
}

func (h *Handler) BuildModeDef(name string) lsha.ModeBuilder {
	mb := &modeBuilder{
		name: name,
	}
	h.modeBuilders[name] = mb
	return mb
}

func (h *Handler) loadPlugins() {
	var orderedPlugins []*pluginBuilder
	remained := map[string]struct{}{}
	for s := range h.pluginBuilders {
		remained[s] = struct{}{}
	}
	loaded := map[string]struct{}{}
	for len(remained) > 0 {
		var thisOrderPlugins []*pluginBuilder
	loadRemain:
		for s := range remained {
			plugin := h.pluginBuilders[s]
			pluginName := plugin.name
			for name, version := range plugin.dependentPluginWithVersion {
				p, ok := h.pluginBuilders[name]
				if !ok {
					h.Warning("dependency %s not found for %s", name, pluginName)
					delete(remained, s)
					continue loadRemain
				}
				if p.version < version {
					h.Warning("dependency %s version %d not satisfied for %s", name, version, pluginName)
					delete(remained, s)
					continue loadRemain
				}
				if _, ok = loaded[s]; !ok {
					continue loadRemain
				}
			}
			thisOrderPlugins = append(thisOrderPlugins, plugin)
			loaded[s] = struct{}{}
			delete(remained, s)
		}
		if len(thisOrderPlugins) > 0 {
			orderedPlugins = append(orderedPlugins, thisOrderPlugins...)
			continue
		}
		for len(remained) > 0 {
			var toCheck string
			for s := range remained {
				toCheck = s
				break
			}
			var path []string
			var findSelf func(p *pluginBuilder) bool
			findSelf = func(p *pluginBuilder) bool {
				if _, ok := p.dependentPluginWithVersion[toCheck]; ok {
					return true
				}
				for name := range p.dependentPluginWithVersion {
					if findSelf(h.pluginBuilders[name]) {
						path = append(path, name)
						return true
					}
				}
				return false
			}
			path = append(path, toCheck)
			findSelf(h.pluginBuilders[toCheck])
			h.Warning("%s dependency cycle: %s -> %s", toCheck, toCheck, strings.Join(path, " -> "))
			delete(remained, toCheck)
		}
	}
	for _, p := range orderedPlugins {
		p.onLoad(h)
	}
	h.validateModes()
}
func (h *Handler) validateModes() {
	for _, builder := range h.modeBuilders {
		switch {
		case builder.start == nil:
			h.ignoreMode(builder, "should have start function")
		case builder.nextTurn == nil:
			h.ignoreMode(builder, "should have next turn function")
		}
	}
}
func (h *Handler) ignoreMode(builder *modeBuilder, reason string) {
	h.Warning("%s %s", builder.name, reason)
	delete(h.modeBuilders, builder.name)
}

func (h *Handler) ListModes(user *User) []string {
	result := make([]string, 0, len(h.modeBuilders))
	for _, m := range h.modeBuilders {
		if m.validateUser(user) == "" {
			result = append(result, m.name)
		}
	}
	sort.Strings(result)
	return result
}
func (h *Handler) CreateRoom(user *User, modeName string) (*Room, error) {
	mb, ok := h.modeBuilders[modeName]
	if !ok {
		return nil, errors.New("mode definition not found: " + modeName)
	}
	if reason := mb.validateUser(user); reason != "" {
		return nil, errors.New("invalidate mode: " + reason)
	}
	if r := user.roomInfo.room.Load(); r != nil {
		return nil, fmt.Errorf("user[%s] has already entered a room: %d", user.id, r.id)
	}
	room := NewRoom(atomic.AddInt64(&h.roomNextID, 1), user, h, mb)
	h.rooms.Store(room.id, room)
	return room, nil
}
func (h *Handler) EnterRoom(user *User, roomID int64) (*Room, error) {
	rawRoom, ok := h.rooms.Load(roomID)
	if !ok {
		return nil, fmt.Errorf("room not found: %v", roomID)
	}
	room := rawRoom.(*Room)
	err := room.Enter(user)
	if err != nil {
		return nil, err
	}
	return room, nil
}
func (h *Handler) RunMode(modeName string, accounts []lsha.User) error {
	mb, ok := h.modeBuilders[modeName]
	if !ok {
		return errors.New("mode definition not found: " + modeName)
	}
	limit := mb.limit
	if limit.UserValidator != nil {
		filtered := make([]lsha.User, 0, len(accounts))
		for _, account := range accounts {
			if limit.UserValidator(account)
		}
	}
	if len(accounts) < limit.PlayerMinCount {
		return errors.New("too less players")
	}
	if limit.PlayerMaxCount > 0 && len(accounts) > limit.PlayerMaxCount {
		accounts = accounts[:limit.PlayerMaxCount]
	}
	ctx := newContext()
	if mb.createConfigFunc != nil {
		ctx.config = mb.createConfigFunc()
		h.AskConfig(ctx.config)
	}
	ctx.mode = mb.start(ctx, accounts)
	for {
		tb := &TurnBuilder{}
		mb.nextTurn(ctx, tb)
		ctx.currentTurn =
	}
	for _, player := range mode.Players() {
		player.Invoke(&lsha.GameStartedEvent{})
	}
	var round int
	for p, newRound := mode.NextPlayer(); p != nil; p, newRound = mode.NextPlayer() {
		if newRound {
			round++
		}
		turn := mode.NewTurn(round, p)
		for phase := turn.NextPhase(); phase != nil; phase = turn.NextPhase() {

			if mode.IsOver() {
				break
			}
		}
	}
}

func (h *Handler) AskConfig(config any) {

}

func (h *Handler) BuildMode(name string) lsha.ModeBuilder {
	if _, ok := h.modeBuilders[name]; ok {
		h.ThrowError(errors.New("duplicate mode definition: " + name))
		return nil
	}
	mb := &modeBuilder{
		name: name,
	}
	h.modeBuilders[name] = mb
	return mb
}
func (h *Handler) Warning(msg string, args ...any) {}
func (h *Handler) ThrowError(err error) {
	// TODO implement me
	panic("implement me")
}
