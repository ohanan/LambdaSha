package basic

import "github.com/ohanan/LambdaSha/pkg/lsha"

func initOneOnOne(mb lsha.ModeBuilder) {
	mb.UserConfig(func(builder lsha.ModeUserConfigBuilder) {
		builder.MaxPlayer(2).MinPlayer(2)
	}).ModeRegistration(func(registration lsha.ModeRegistration) {

	}).Init(func(ctx lsha.Context, userBuilders []lsha.ModeInitUserBuilder) (ctxData any) {
		mode := &oneOnOne{}
		for _, builder := range userBuilders {
			builder.BindData(&oneOnOnePlayer{})
		}
		return mode
	}).NextTurn(nextTurn)
}

type OneOnOneMode interface {
	Mode
}

type OneOnOnePlayer interface {
}
type OneOnOneTurn interface {
}
type OneOnOnePhase interface {
}
type oneOnOne struct {
}

func (o *oneOnOne) Init() any {
	return o
}

type oneOnOnePlayer struct {
}
type oneOnOneTurn struct {
}
