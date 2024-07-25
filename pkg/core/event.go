package core

import (
	"github.com/ohanan/LambdaSha/pkg/lsha"
)

type invokerResult struct {
}

func (i *invokerResult) FastStop() {
}

type Trigger struct {
	id uint64
	lsha.Trigger
	player       lsha.Player
	eventNameMap map[string]struct{}
}

func (t *Trigger) getTriggerPlayerOrder() int {
	if t.player == nil {
		return -1
	}
	return t.player.Order()
}
