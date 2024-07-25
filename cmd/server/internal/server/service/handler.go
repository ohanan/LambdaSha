package service

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/ohanan/LambdaSha/pkg/core"
	"github.com/ohanan/LambdaSha/pkg/lsha"
)

var handlerInitOnce sync.Once
var handler *Handler

func GetHandler() *Handler {
	handlerInitOnce.Do(func() {
		handler = &Handler{
			modeBuilders:   map[string]core.BuiltMode{},
			pluginBuilders: map[string]core.BuiltPlugin{},
		}
	})
	return handler
}

type Handler struct {
	modeBuilders   map[string]core.BuiltMode
	pluginBuilders map[string]core.BuiltPlugin

	rooms      sync.Map
	roomNextID int64
}

func (h *Handler) Init(plugins map[string]lsha.PluginRegister) {
	for name, initializer := range plugins {
		h.pluginBuilders[name] = core.BuildPlugin(initializer)
	}
	h.loadPlugins()
}

func (h *Handler) GetModeRegistration(name string) lsha.ModeRegistration {
	return h.modeBuilders[name]
}

func (h *Handler) BuildMode(f func(builder lsha.ModeBuilder)) {
	mode := core.BuildMode(f)
	name := mode.GetName()
	if _, ok := h.modeBuilders[name]; ok {
		h.ThrowError(errors.New("duplicate mode definition: " + name))
		return
	}
	h.modeBuilders[name] = mode
}

func (h *Handler) loadPlugins() {
	var orderedPlugins []core.BuiltPlugin
	remained := map[string]struct{}{}
	for s := range h.pluginBuilders {
		remained[s] = struct{}{}
	}
	loaded := map[string]struct{}{}
	for len(remained) > 0 {
		var thisOrderPlugins []core.BuiltPlugin
	loadRemain:
		for s := range remained {
			plugin := h.pluginBuilders[s]
			pluginName := plugin.GetName()
			for name, version := range plugin.GetDependents() {
				p, ok := h.pluginBuilders[name]
				if !ok {
					h.Warning("dependency %s not found for %s", name, pluginName)
					delete(remained, s)
					continue loadRemain
				}
				if p.GetVersion() < version {
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
			var findSelf func(p core.BuiltPlugin) bool
			findSelf = func(p core.BuiltPlugin) bool {
				if _, ok := p.GetDependents()[toCheck]; ok {
					return true
				}
				for name := range p.GetDependents() {
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
		p.Load(h)
	}
}

func (h *Handler) ListModes(user *User) []string {
	result := make([]string, 0, len(h.modeBuilders))
	for _, m := range h.modeBuilders {
		if m.ValidateUser(user) == "" {
			result = append(result, m.GetName())
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
	if reason := mb.ValidateUser(user); reason != "" {
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

func (h *Handler) Warning(msg string, args ...any) {}
func (h *Handler) ThrowError(err error) {
	// TODO implement me
	panic("implement me")
}
