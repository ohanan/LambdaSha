package core

import (
	"github.com/ohanan/LambdaSha/pkg/core/form"
	"github.com/ohanan/LambdaSha/pkg/lsha"
)

type modeBuilder struct {
	name            string
	limit           *lsha.ModeLimit
	description     string
	buildConfigFunc lsha.ModeRoomConfigBuilder
	start           lsha.ModeStarter
	nextTurn        lsha.TurnStarter
}

func (b *modeBuilder) createConfigPointer() *form.ItemsBuilder {
	v := &form.ItemsBuilder{}
	if b.buildConfigFunc != nil {
		b.buildConfigFunc(v)
	}
	return v
}
func (b *modeBuilder) validateUser(user *User) (reason string) {
	l := b.limit
	if l == nil {
		return ""
	}
	v := l.UserValidator
	if v == nil {
		return ""
	}
	return v(user)
}
func (b *modeBuilder) getMaxPlayerCount() int {
	l := b.limit
	if l == nil {
		return defaultMaxPlayerCount
	}
	maxCount := l.PlayerMaxCount
	if maxCount < 1 {
		return 1
	}
	return maxCount
}

func (b *modeBuilder) isTooMuchPlayers(userCount int) bool {
	l := b.limit
	if l == nil {
		return false
	}
	return l.PlayerMaxCount > 0 && l.PlayerMaxCount < userCount
}
func (b *modeBuilder) isTooLessPlayers(userCount int) bool {
	l := b.limit
	if l == nil {
		return false
	}
	return l.PlayerMinCount > 0 && l.PlayerMinCount > userCount
}
func (b *modeBuilder) Limit(limit *lsha.ModeLimit) {
	b.limit = limit
}

func (b *modeBuilder) OnCreateConfig(f lsha.ModeRoomConfigBuilder) {
	b.buildConfigFunc = f
}

func (b *modeBuilder) OnStart(f lsha.ModeStarter) {
	b.start = f
}

func (b *modeBuilder) OnNextTurn(f lsha.TurnStarter) {
	b.nextTurn = f
}
func (b *modeBuilder) WithModeRegistration(f func(registration lsha.ModeRegistration)) {
	f(b)
}

func (b *modeBuilder) OnEvent(ctx lsha.Context, e lsha.Event, result lsha.Event) {
	// TODO implement me
	panic("implement me")
}

func (b *modeBuilder) SetHeroDef(h lsha.HeroDef) {
	// TODO implement me
	panic("implement me")
}

func (b *modeBuilder) DeleteHeroDef(name string) {
	// TODO implement me
	panic("implement me")
}

func (b *modeBuilder) Description(description string) {
	b.description = description
}
