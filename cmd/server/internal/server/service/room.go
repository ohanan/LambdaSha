package service

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/ohanan/LambdaSha/pkg/core"
	"github.com/ohanan/LambdaSha/pkg/core/common"
	"github.com/ohanan/LambdaSha/pkg/core/form"
	"github.com/ohanan/LambdaSha/pkg/lsha"
)

const defaultMaxPlayerCount = 32

func NewRoom(id int64, user *User, e *Handler, mode core.BuiltMode) *Room {
	r := &Room{
		id: id,
		h:  e,
	}
	if rr := user.roomInfo.room.Load(); rr != nil {
		_ = rr.Leave(user)
	}
	defer user.roomInfo.room.Store(r)
	r.mode.Store(&mode)
	r.resetConfigForMode(mode)
	r.owner.Store(user)
	_, maxPlayer := mode.GetPlayerCountLimit()
	users := make([]*User, maxPlayer)
	users[0] = user
	r.users.Store(&users)
	spectators := make([]*User, 0, defaultMaxPlayerCount)
	r.spectators.Store(&spectators)
	return r
}

type Room struct {
	id              int64
	name            atomic.Pointer[string]
	mode            atomic.Pointer[core.BuiltMode]
	ownerReadConfig atomic.Pointer[[]*form.Item]
	configFactory   atomic.Pointer[func(readonly bool) []*form.Item]
	configData      atomic.Pointer[any]
	owner           atomic.Pointer[User]
	users           atomic.Pointer[[]*User]
	spectators      atomic.Pointer[[]*User]
	sync.RWMutex

	h *Handler
}

func (r *Room) Enter(user *User) error {
	userRoomInfo := user.roomInfo
	userRoomInfo.Lock()
	defer userRoomInfo.Unlock()
	if rr := userRoomInfo.room.Load(); rr != nil {
		if rr.id == r.id {
			return nil
		}
		return fmt.Errorf("user[%s] has already entered a room: %d", user.id, r.id)
	}
	r.Lock()
	defer r.Unlock()
	m := *r.mode.Load()
	if reason := m.ValidateUser(user); reason != "" {
		return r.addToSpectators(user)
	}
	players := *r.users.Load()
	for i, player := range players {
		if player == nil {
			players[i] = user
			userRoomInfo.isSpectator.Store(false)
			return nil
		}
	}
	return r.addToSpectators(user)
}

func (r *Room) Rename(v string) {
	r.name.Store(&v)
}
func (r *Room) Leave(user *User) error {
	userRoomInfo := user.roomInfo
	userRoomInfo.Lock()
	defer userRoomInfo.Unlock()
	if rr := userRoomInfo.room.Load(); rr == nil {
		return fmt.Errorf("user[%s] is not in any room", user.id)
	} else if rr.id != r.id {
		return fmt.Errorf("user[%s] is in other room: %d, but not: %d", user.id, rr.id, r.id)
	}
	userRoomInfo.isSpectator.Store(false)
	defer userRoomInfo.room.Store(nil)
	r.Lock()
	defer r.Unlock()
	if userRoomInfo.isSpectator.Load() {
		load := *r.spectators.Load()
		for id, u := range load {
			if u.id == user.id {
				load = append(load[:id], load[id+1:]...)
				r.spectators.Store(&load)
				return nil // fast return if user is in spectators
			}
		}
	}
	players := *r.users.Load()
	for i, player := range players { // remove user from users
		if player != nil && player.id == user.id {
			players[i] = nil
			break
		}
	}
	idx, firstUser := common.FirstNotNil(players)
	if idx < 0 {
		r.h.rooms.Delete(r.id)
		return nil
	}
	if r.owner.Load().id == user.id {
		r.owner.Store(firstUser)
	}
	return nil
}

func (r *Room) SetMode(user *User, modeName string) (err error) {
	mode, ok := r.h.modeBuilders[modeName]
	if !ok {
		return fmt.Errorf("no such mode: %s", modeName)
	}
	if err = r.mustBeOwner(user); err != nil {
		return err
	}
	r.Lock()
	defer r.Unlock()
	if err = r.mustBeOwner(user); err != nil {
		return err
	}
	curr := *r.mode.Load()
	if curr.GetName() == mode.GetName() {
		return nil
	}
	defer r.mode.Store(&mode)
	r.resetConfigForMode(mode)
	players := *r.users.Load()
	_, newSize := mode.GetPlayerCountLimit()
	newPlayers := make([]*User, newSize)
	r.users.Store(&newPlayers)
	if len(players) < newSize {
		copy(newPlayers, players)
		return nil
	}
	var lastNilIdx int
	for _, player := range players {
		if player != nil {
			players[lastNilIdx] = player
			lastNilIdx++
		}
	}
	players = players[:lastNilIdx] // remove nil
	copy(newPlayers, players)
	if lastNilIdx <= newSize {
		return nil
	}
	for i, player := range players {
		if player.id == r.owner.Load().id {
			if i >= newSize {
				players[0], players[i] = players[i], players[0]
			}
			break
		}
	}

	for i := lastNilIdx; i < len(players); i++ {
		user := players[i]
		_ = r.addToSpectators(user)
		userRoomInfo := user.roomInfo
		userRoomInfo.isSpectator.Store(true)
		userRoomInfo.Unlock()
	}
	return nil
}
func (r *Room) ResetConfig() {
	r.Lock()
	defer r.Unlock()
	r.resetConfigForMode(*r.mode.Load())
}
func (r *Room) resetConfigForMode(m core.BuiltMode) {
	data, creator := m.CreateConfigBuilder()
	r.configData.Store(&data)
	r.configFactory.Store(&creator)
}

func (r *Room) GetConfig(user *User) []*form.Item {
	ownerReadConfig := r.owner.Load().ID() != user.ID()
	items := (*r.configFactory.Load())(ownerReadConfig)
	if ownerReadConfig {
		r.ownerReadConfig.Store(&items)
	}
	return items
}
func (r *Room) UpdateConfig(user *User, updateContent map[string]any) error {
	if err := r.mustBeOwner(user); err != nil {
		return err
	}
	ownerReadConfig := r.ownerReadConfig.Load()
	if ownerReadConfig != nil {
		form.UpdateItems(*ownerReadConfig, updateContent)
	}
	return nil
}
func (r *Room) Start() error {
	r.Lock()
	defer r.Unlock()
	mode := *r.mode.Load()
	users := *r.users.Load()
	minCount, _ := mode.GetPlayerCountLimit()
	if currentPlayerCnt := len(users); currentPlayerCnt < minCount {
		return fmt.Errorf("not enough users, expected: %d, got: %d", minCount, currentPlayerCnt)
	}
	copiedUsers := make([]lsha.User, len(users))
	for i, user := range users {
		copiedUsers[i] = user
	}
	go mode.Run(*r.configData.Load(), copiedUsers)
	return nil
}

func (r *Room) mustBeOwner(user *User) error {
	if r.owner.Load().id == user.id {
		return nil
	}
	return fmt.Errorf("user[%s] is not owner of room: %d", user.id, r.id)
}

// addToSpectators add user to spectators, it is not goroutine safe.
func (r *Room) addToSpectators(user *User) error {
	common.AtomicAppend(&r.spectators, user)
	user.roomInfo.isSpectator.Store(true)
	return nil
}

func (r *Room) doEnter(user *User) error {
	user.roomInfo.room.Store(r)
	ps := *r.users.Load()
	ps = append(ps, user)
	r.users.Store(&ps)
	return nil
}

type UserRoomInfo struct {
	room        atomic.Pointer[Room]
	isSpectator atomic.Bool
	sync.Mutex
}
