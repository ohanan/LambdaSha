package core

import (
	"github.com/ohanan/LambdaSha/pkg/lsha"
)

type modeBuilder struct {
	name             string
	limit            *lsha.ModeLimit
	description      string
	createConfigFunc lsha.FuncModeCreateConfig
	start            lsha.FuncModeStart
	nextTurn         lsha.FuncModeNextTurn
}

func (b *modeBuilder) createConfigPointer() *any {
	if b.createConfigFunc != nil {
		c := b.createConfigFunc()
		return &c
	}
	return nil
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

func (b *modeBuilder) OnCreateConfig(f lsha.FuncModeCreateConfig) {
	b.createConfigFunc = f
}

func (b *modeBuilder) OnStart(f lsha.FuncModeStart) {
	b.start = f
}

func (b *modeBuilder) OnNextTurn(f lsha.FuncModeNextTurn) {
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
