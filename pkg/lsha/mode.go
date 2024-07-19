package lsha

type (
	ModeConfig           = any
	Mode                 = any
	FuncModeCreateConfig = func() (config ModeConfig)
	FuncModeStart        = func(ctx Context) Mode
	FuncModeNextTurn     = func(ctx Context, turnBuilder TurnBuilder)
	FuncModeNextPlayer   = func(ctx Context, player Player)
)
type ModeRepository interface {
	GetModeRegistration(name string) ModeRegistration
	BuildModeDef(name string) ModeBuilder
}

type ModeRegistration interface {
	SetHeroDef(h HeroDef)
	DeleteHeroDef(name string)
}

type ModeLimit struct {
	PlayerMinCount int
	PlayerMaxCount int
	UserValidator  func(account User) (reason string)
}
type ModeBuilder interface {
	Description(description string)
	Limit(limit *ModeLimit)
	WithModeRegistration(f func(registration ModeRegistration))
	OnCreateConfig(f FuncModeCreateConfig)
	OnStart(f FuncModeStart)
	OnNextTurn(f FuncModeNextTurn)
	OnEvent(ctx Context, e Event, result Event)
}
