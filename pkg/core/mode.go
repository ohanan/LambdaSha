package core

import (
	"math/rand"

	"github.com/ohanan/LambdaSha/pkg/core/common"
	"github.com/ohanan/LambdaSha/pkg/core/form"
	"github.com/ohanan/LambdaSha/pkg/lsha"
)

type BuiltMode interface {
	lsha.ModeRegistration
	GetName() string
	GetDescription() string
	GetPlayerCountLimit() (min, max int)
	ValidateUser(user lsha.User) (reason string)
	CreateConfigBuilder() (configData any, creator func(readonly bool) []*form.Item)
	Run(configData any, users []lsha.User)
}

func BuildMode(r func(lsha.ModeBuilder)) BuiltMode {
	p := newModeBuilder()
	r(p)
	return p
}

func newModeBuilder() *modeBuilder {
	return &modeBuilder{
		userConfig: &modeConfigBuilder{
			playerMinCount: 1,
			playerMaxCount: 16,
			userValidator:  noUserCheck,
		},
		description:     "no description",
		buildConfigFunc: func(roomConfigBuilder lsha.ConfigBuilder) {},
		initializer:     func(ctx lsha.Context, builders []lsha.ModeInitUserBuilder) (ctxData any) { return nil },
		nextTurn:        func(ctx lsha.Context, tb lsha.TurnBuilder) (turnData any) { return nil },
	}
}

type modeBuilder struct {
	name            string
	userConfig      *modeConfigBuilder
	description     string
	buildConfigFunc lsha.ModeRoomConfigBuilder
	initializer     lsha.ModeInitializer
	nextTurn        lsha.TurnStarter
}

func (b *modeBuilder) GetName() string {
	return b.name
}
func (b *modeBuilder) GetPlayerCountLimit() (min, max int) {
	return b.userConfig.playerMinCount, b.userConfig.playerMaxCount
}
func (b *modeBuilder) ValidateUser(user lsha.User) (reason string) {
	return b.userConfig.userValidator(user)
}
func (b *modeBuilder) GetDescription() string {
	return b.description
}

func (b *modeBuilder) CreateConfigBuilder() (configData any, creator func(readonly bool) []*form.Item) {
	f := &form.ItemsBuilder{}
	b.buildConfigFunc(f)
	return f.Data(), f.Build
}

func (b *modeBuilder) Name(name string) lsha.ModeBuilder {
	b.name = name
	return b
}
func (b *modeBuilder) BuildConfig() *form.ItemsBuilder {
	v := &form.ItemsBuilder{}
	if b.buildConfigFunc != nil {
		b.buildConfigFunc(v)
	}
	return v
}

func (b *modeBuilder) UserConfig(builder func(builder lsha.ModeUserConfigBuilder)) lsha.ModeBuilder {
	builder(b.userConfig)
	return b
}
func noUserCheck(user lsha.User) (reason string) {
	return ""
}

func (b *modeBuilder) OnCreateConfig(f lsha.ModeRoomConfigBuilder) lsha.ModeBuilder {
	if f != nil {
		b.buildConfigFunc = f
	}
	return b
}

func (b *modeBuilder) Init(f lsha.ModeInitializer) lsha.ModeBuilder {
	if f != nil {
		b.initializer = f
	}
	return b
}

func (b *modeBuilder) NextTurn(f lsha.TurnStarter) lsha.ModeBuilder {
	if f != nil {
		b.nextTurn = f
	}
	return b
}
func (b *modeBuilder) ModeRegistration(f func(registration lsha.ModeRegistration)) lsha.ModeBuilder {
	if f != nil {
		f(b)
	}
	return b
}
func (b *modeBuilder) SetHeroDef(h lsha.HeroDef) {
	// TODO implement me
	panic("implement me")
}

func (b *modeBuilder) DeleteHeroDef(name string) {
	// TODO implement me
	panic("implement me")
}

func (b *modeBuilder) Description(description string) lsha.ModeBuilder {
	b.description = description
	return b
}
func (b *modeBuilder) Run(configData any, users []lsha.User) {
	ctx := newContext(b, configData, users)
	{
		copied := make([]lsha.User, len(users))
		copy(copied, users)
		users = copied
	}
	if !b.userConfig.disableRandomOrder {
		rand.Shuffle(len(users), func(i, j int) {
			users[i], users[j] = users[j], users[i]
		})
	}

	initBuilders := make([]lsha.ModeInitUserBuilder, len(users))
	for i, user := range users {
		initBuilders[i] = &ModeInitUserBuilder{
			user:  user,
			order: i,
		}
	}
	ctx.data.Store(common.Ptr(b.initializer(ctx, initBuilders)))
	players := make([]*Player, len(users))
	for i, builder := range initBuilders {
		b := builder.(*ModeInitUserBuilder)
		players[i] = &Player{
			data:  b.data,
			order: b.order,
			user:  b.user,
		}
	}
	ctx.players.Store(common.Ptr(players))
	for _, player := range players {
		event := &lsha.PlayerPreparedEvent{}
		event.SetPlayer(player)
		ctx.Invoke(event)
	}
	ctx.Invoke(&lsha.GameStartedEvent{})
	for i := 2024; i > 0; i-- {
		lastTurn := ctx.turn.Load()
		tb := &TurnBuilder{}
		turn := &Turn{}
		turn.data = b.nextTurn(ctx, tb)
		if tb.player == nil {
			return
		}
		turn.player = tb.player
		turn.round = tb.round
		if turn.round <= 0 {
			turn.round = lastTurn.round + 1
		}
		ctx.turn.Store(turn)
		turnStartedEvent := &lsha.TurnStartedEvent{}
		ctx.Invoke(turnStartedEvent)
		for j := 100; j > 0; j-- {
			pb := &PhaseBuilder{}
			phase := &Phase{}
			phase.data = tb.nextPhase(ctx, pb)
			if pb.name == "" {
				break
			}
			phase.name = pb.name
			turn.phase = phase
			phaseStartedEvent := &lsha.PhaseStartedEvent{}
			phaseStartedEvent.SetPhase(phase)
			phaseStartedEvent.SetTurn(turn)
			ctx.Invoke(phaseStartedEvent)
		}
	}
}

type modeConfigBuilder struct {
	playerMinCount     int
	playerMaxCount     int
	userValidator      func(account lsha.User) (reason string)
	disableRandomOrder bool
}

func (m *modeConfigBuilder) MinPlayer(playerCount int) lsha.ModeUserConfigBuilder {
	if playerCount > 0 {
		m.playerMinCount = playerCount
		if m.playerMaxCount < m.playerMinCount {
			m.playerMaxCount = m.playerMinCount
		}
	}
	return m
}

func (m *modeConfigBuilder) MaxPlayer(playerCount int) lsha.ModeUserConfigBuilder {
	if playerCount > m.playerMinCount {
		m.playerMaxCount = playerCount
	}
	return m
}

func (m *modeConfigBuilder) ValidUser(validator func(user lsha.User) (reason string)) lsha.ModeUserConfigBuilder {
	if validator != nil {
		m.userValidator = validator
	}
	return m
}
func (m *modeConfigBuilder) DisableRandomOrder() lsha.ModeUserConfigBuilder {
	m.disableRandomOrder = true
	return m
}

type ModeInitUserBuilder struct {
	user  lsha.User
	order int
	data  any
}

func (m *ModeInitUserBuilder) User() lsha.User {
	return m.user
}

func (m *ModeInitUserBuilder) Order() int {
	return m.order
}

func (m *ModeInitUserBuilder) RewriteOrder(order int) lsha.ModeInitUserBuilder {
	m.order = order
	return m
}

func (m *ModeInitUserBuilder) BindData(data any) lsha.ModeInitUserBuilder {
	m.data = data
	return m
}
