package lsha

type (
	ModeRoomConfigBuilder = func(roomConfigBuilder ConfigBuilder)
	ModeStarter           = func(ctx Context) (ctxData any)
	FuncModeNextTurn      = func(ctx Context, turnBuilder TurnBuilder)
	FuncModeNextPlayer    = func(ctx Context, player Player)
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
	OnCreateConfig(f ModeRoomConfigBuilder)
	OnStart(f ModeStarter)
	OnNextTurn(f TurnStarter)
	OnEvent(ctx Context, e Event, result Event)
}
type ConfigBuilder interface {
	BindData(data any)
	Data() any
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
