package core

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/ohanan/LambdaSha/pkg/core/form"
)

const defaultMaxPlayerCount = 32

func NewRoom(id int64, user *User, e *Handler, mode *modeBuilder) *Room {
	r := &Room{
		id: id,
		h:  e,
	}
	if rr := user.roomInfo.room.Load(); rr != nil {
		_ = rr.Leave(user)
	}
	defer user.roomInfo.room.Store(r)
	r.config.Store(mode.createConfigPointer())
	r.mode.Store(mode)
	r.owner.Store(user)
	users := make([]*User, mode.getMaxPlayerCount())
	users[0] = user
	r.players.Store(&users)
	spectators := make([]*User, 0, defaultMaxPlayerCount)
	r.spectators.Store(&spectators)
	return r
}

type Room struct {
	id         int64
	name       atomic.Pointer[string]
	mode       atomic.Pointer[modeBuilder]
	owner      atomic.Pointer[User]
	players    atomic.Pointer[[]*User]
	spectators atomic.Pointer[[]*User]
	config     atomic.Pointer[form.ItemsBuilder]
	ctx        atomic.Pointer[runtimeContext]
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
	m := r.mode.Load()
	if reason := m.validateUser(user); reason != "" {
		return r.addToSpectators(user)
	}
	players := *r.players.Load()
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
	players := *r.players.Load()
	for i, player := range players { // remove user from players
		if player != nil && player.id == user.id {
			players[i] = nil
			break
		}
	}
	idx, firstUser := firstNotNil(players)
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
	curr := r.mode.Load()
	if curr.name == mode.name {
		return nil
	}
	defer r.mode.Store(mode)
	r.config.Store(mode.createConfigPointer())
	players := *r.players.Load()
	newSize := mode.getMaxPlayerCount()
	newPlayers := make([]*User, newSize)
	r.players.Store(&newPlayers)
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
	r.config.Store(r.mode.Load().createConfigPointer())
}
func (r *Room) GetConfig(user *User) []*form.Item {
	return (*r.config.Load()).Build(r.owner.Load().ID() != user.ID())
}
func (r *Room) Start() error {
	r.Lock()
	defer r.Unlock()
	mode := r.mode.Load()
	currentPlayerCnt := len(*r.players.Load())
	if mode.limit != nil {
		if currentPlayerCnt < mode.limit.PlayerMinCount {
			return fmt.Errorf("not enough players, expected: %d, got: %d", mode.limit.PlayerMinCount, currentPlayerCnt)
		}
	}
	ctx := newContext(r.config.Load(), r.players.Load())
	go r.run(ctx, mode)
	return nil
}
func (r *Room) run(ctx *Context, mode *modeBuilder) {
	ctx.data.Store(ptr(mode.start(ctx)))
	for {
		lastTurn := ctx.turn.Load()
		tb := &TurnBuilder{}
		turn := &Turn{}
		turn.data = mode.nextTurn(ctx, tb)
		if tb.player == nil {
			return
		}
		turn.player = tb.player
		turn.round = tb.round
		if turn.round == 0 {
			turn.round = lastTurn.round + 1
		}
		ctx.turn.Store(turn)
		for {
			pb := &PhaseBuilder{}
			phase := &Phase{}
			phase.data = tb.nextPhase(ctx, pb)
			if pb.name == "" {
				break
			}
			phase.name = pb.name
			turn.phase = phase
		}
	}
}

func (r *Room) mustBeOwner(user *User) error {
	if r.owner.Load().id == user.id {
		return nil
	}
	return fmt.Errorf("user[%s] is not owner of room: %d", user.id, r.id)
}

// addToSpectators add user to spectators, it is not goroutine safe.
func (r *Room) addToSpectators(user *User) error {
	atomicAppend(&r.spectators, user)
	user.roomInfo.isSpectator.Store(true)
	return nil
}

func (r *Room) doEnter(user *User) error {
	user.roomInfo.room.Store(r)
	ps := *r.players.Load()
	ps = append(ps, user)
	r.players.Store(&ps)
	return nil
}

type UserRoomInfo struct {
	room        atomic.Pointer[Room]
	isSpectator atomic.Bool
	sync.Mutex
}
