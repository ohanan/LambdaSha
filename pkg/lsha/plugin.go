package lsha

type PluginBuilder interface {
	Name(name string)
	Description(description string)
	Version(v int)
	Dependencies(dependentPluginWithVersion map[string]int)
	OnLoad(f func(repository ModeRepository))
}
