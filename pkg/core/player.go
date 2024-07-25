package core

import (
	"github.com/ohanan/LambdaSha/pkg/lsha"
)

type Player struct {
	data  any
	order int
	user  lsha.User
	dead  bool
}

func (p *Player) BindData(data any) {
	p.data = data
}

func (p *Player) Data() any {
	return p.data
}

func (p *Player) Order() int {
	return p.order
}

func (p *Player) User() lsha.User {
	return p.user
}

func (p *Player) IsAlive() bool {
	return !p.dead
}

func (p *Player) Effects() lsha.Effect {
	// TODO implement me
	panic("implement me")
}
