package lsha

type (
	PluginRegister func(pb PluginBuilder)
)
type PluginBuilder interface {
	Name(name string) PluginBuilder
	Description(d string) PluginBuilder
	Version(v int) PluginBuilder
	Dependencies(dependentPluginWithVersion map[string]int) PluginBuilder
	OnLoad(f func(r ModeRepository)) PluginBuilder
}
