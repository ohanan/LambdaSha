package lsha

type (
	ModeRoomConfigBuilder = func(roomConfigBuilder ConfigBuilder)
	ModeInitializer       = func(ctx Context, userBuilders []ModeInitUserBuilder) (ctxData any)
	PrepareUser           = func(order int, users User) (playerData any)
	FuncModeNextTurn      = func(ctx Context, turnBuilder TurnBuilder)
)
type ModeRepository interface {
	GetModeRegistration(name string) ModeRegistration
	BuildMode(f func(builder ModeBuilder))
}

type ModeRegistration interface {
	SetHeroDef(h HeroDef)
	DeleteHeroDef(name string)
}
type ModeInitUserBuilder interface {
	User() User
	Order() int
	RewriteOrder(order int) ModeInitUserBuilder
	BindData(data any) ModeInitUserBuilder
}
type ModeUserConfigBuilder interface {
	MinPlayer(playerCount int) ModeUserConfigBuilder
	MaxPlayer(playerCount int) ModeUserConfigBuilder
	ValidUser(validator func(user User) (reason string)) ModeUserConfigBuilder
	DisableRandomOrder() ModeUserConfigBuilder
}
type ModeBuilder interface {
	Name(name string) ModeBuilder
	Description(description string) ModeBuilder
	UserConfig(builderInitializer func(builder ModeUserConfigBuilder)) ModeBuilder
	ModeRegistration(f func(registration ModeRegistration)) ModeBuilder
	OnCreateConfig(f ModeRoomConfigBuilder) ModeBuilder
	Init(f ModeInitializer) ModeBuilder
	NextTurn(f TurnStarter) ModeBuilder
}
type ConfigBuilder interface {
	DataHolder
	Desc(desc string) ConfigDescOptionsBuilder
	Checkbox(name string, tips string) ConfigCheckboxOptionsBuilder
	Radio(name string, tips string) ConfigRadioOptionsBuilder
	Range(name string, tips string) ConfigRangeOptionsBuilder
}
type ConfigOptionsBuilder interface {
	Parent() ConfigBuilder
}
type ConfigDescOptionsBuilder interface {
	ConfigOptionsBuilder
	Desc(desc string) ConfigDescOptionsBuilder
}
type ConfigCheckboxOptionsBuilder interface {
	ConfigOptionsBuilder
	SetName(name string) ConfigCheckboxOptionsBuilder
	SetTips(tips string) ConfigCheckboxOptionsBuilder
	AddOption(name string, tips string, checked bool, onChanged func(data any, name, checkboxName string, checked bool)) ConfigCheckboxOptionsBuilder
	ResetOptions() ConfigCheckboxOptionsBuilder
	RemoveOption(name string) ConfigCheckboxOptionsBuilder
	SetOptionTips(name string, tips string) ConfigCheckboxOptionsBuilder
	CheckOption(name string, checked bool) ConfigCheckboxOptionsBuilder
	OnChangedOption(name string, onChanged func(data any, name, checkboxName string, checked bool)) ConfigCheckboxOptionsBuilder
}
type ConfigRadioOptionsBuilder interface {
	ConfigOptionsBuilder
	SetName(name string) ConfigRadioOptionsBuilder
	SetTips(tips string) ConfigRadioOptionsBuilder
	AddOption(name string, tips string) ConfigRadioOptionsBuilder
	CheckOption(name string) ConfigRadioOptionsBuilder
	ResetOptions() ConfigRadioOptionsBuilder
	RemoveOption(name string) ConfigRadioOptionsBuilder
	OnCheckedOption(onChecked func(data any, name, radioName string)) ConfigRadioOptionsBuilder
	Parent() ConfigBuilder
}
type ConfigRangeOptionsBuilder interface {
	ConfigOptionsBuilder
	SetName(name string) ConfigRangeOptionsBuilder
	SetTips(tips string) ConfigRangeOptionsBuilder
	Max(value int, label string) ConfigRangeOptionsBuilder
	Min(value int, label string) ConfigRangeOptionsBuilder
	ValueTips(value int, tips string) ConfigRangeOptionsBuilder
	Value(value int) ConfigRangeOptionsBuilder
	OnChanged(onChanged func(data any, name string, value int)) ConfigRangeOptionsBuilder
	RemoveValueTip(value int) ConfigRangeOptionsBuilder
}
