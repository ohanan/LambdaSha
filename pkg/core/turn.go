package core

import (
	"github.com/ohanan/LambdaSha/pkg/lsha"
)

var _ lsha.TurnBuilder = (*TurnBuilder)(nil)

type Turn struct {
	data   any
	player lsha.Player
	round  int
	phase  *Phase
}

func (t *Turn) BindData(data any) {
	t.data = data
}

func (t *Turn) Data() any {
	return t.data
}

func (t *Turn) Player() lsha.Player {
	return t.player
}

func (t *Turn) Round() int {
	return t.round
}

func (t *Turn) Phase() lsha.Phase {
	return t.phase
}

type TurnBuilder struct {
	round     int
	player    lsha.Player
	nextPhase lsha.PhaseStarter
}

func (t *TurnBuilder) OnNextPhase(phaseStarter lsha.PhaseStarter) lsha.TurnBuilder {
	t.nextPhase = phaseStarter
	return t
}

func (t *TurnBuilder) Player(p lsha.Player) lsha.TurnBuilder {
	t.player = p
	return t
}

func (t *TurnBuilder) Round(n int) lsha.TurnBuilder {
	t.round = n
	return t
}
