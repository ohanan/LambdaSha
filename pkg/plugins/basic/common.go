package basic

import "github.com/ohanan/LambdaSha/pkg/lsha"

type Mode struct {
}
type Player struct {
}
type Turn struct {
}
type Phase interface {
	NextPhase() Phase
	Name() string
}

type StartPhase struct {
}

type PreCheckPhase struct {
}

type HarvestPhase struct {
}
type PlayPhase struct {
}
type PostCheckPhase struct {
}
type EndPhase struct {
}

func (s *StartPhase) NextPhase() Phase     { return &PreCheckPhase{} }
func (s *PreCheckPhase) NextPhase() Phase  { return &HarvestPhase{} }
func (s *HarvestPhase) NextPhase() Phase   { return &PlayPhase{} }
func (s *PlayPhase) NextPhase() Phase      { return &PostCheckPhase{} }
func (s *PostCheckPhase) NextPhase() Phase { return &EndPhase{} }
func (s *EndPhase) NextPhase() Phase       { return nil }

func (s *StartPhase) Name() string     { return PhaseStart }
func (s *PreCheckPhase) Name() string  { return PhasePreCheck }
func (s *HarvestPhase) Name() string   { return PhaseHarvest }
func (s *PlayPhase) Name() string      { return PhasePlay }
func (s *PostCheckPhase) Name() string { return PhasePostCheck }
func (s *EndPhase) Name() string       { return PhaseEnd }

func nextTurn(ctx lsha.Context, tb lsha.TurnBuilder) (turnData any) {
	tb.Player(ctx.NextPlayer(nil)).OnNextPhase(func(ctx lsha.Context, pb lsha.PhaseBuilder) (phaseData any) {
		phase := lsha.PhaseData[Phase](ctx)
		if phase == nil {
			phase = &StartPhase{}
		}
		pb.Name(phase.Name())
		return phase
	})
	turn := &oneOnOneTurn{}
	return turn
}
