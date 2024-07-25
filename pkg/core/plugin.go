package core

import (
	"github.com/ohanan/LambdaSha/pkg/lsha"
)

func BuildPlugin(r lsha.PluginRegister) BuiltPlugin {
	p := &pluginBuilder{}
	r(p)
	return p
}

type BuiltPlugin interface {
	GetName() string
	GetDescription() string
	GetVersion() int
	GetDependents() map[string]int
	Load(registration lsha.ModeRepository)
}
type pluginBuilder struct {
	name                       string
	description                string
	version                    int
	dependentPluginWithVersion map[string]int
	onLoad                     func(registration lsha.ModeRepository)
}

func (p *pluginBuilder) Name(name string) lsha.PluginBuilder {
	p.name = name
	return p
}

func (p *pluginBuilder) Description(description string) lsha.PluginBuilder {
	p.description = description
	return p
}

func (p *pluginBuilder) Version(v int) lsha.PluginBuilder {
	p.version = v
	return p
}

func (p *pluginBuilder) Dependencies(dependentPluginWithVersion map[string]int) lsha.PluginBuilder {
	p.dependentPluginWithVersion = dependentPluginWithVersion
	return p
}

func (p *pluginBuilder) OnLoad(f func(registration lsha.ModeRepository)) lsha.PluginBuilder {
	p.onLoad = f
	return p
}

func (p *pluginBuilder) GetName() string {
	return p.name
}
func (p *pluginBuilder) GetDescription() string {
	return p.description
}
func (p *pluginBuilder) GetVersion() int {
	return p.version
}
func (p *pluginBuilder) GetDependents() map[string]int {
	return p.dependentPluginWithVersion
}
func (p *pluginBuilder) Load(registration lsha.ModeRepository) {
	if p.onLoad != nil {
		p.onLoad(registration)
	}
}
