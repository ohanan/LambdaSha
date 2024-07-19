package core

import (
	"github.com/ohanan/LambdaSha/pkg/lsha"
)

type pluginBuilder struct {
	name                       string
	description                string
	version                    int
	dependentPluginWithVersion map[string]int
	onLoad                     func(registration lsha.ModeRepository)
}

func (p *pluginBuilder) Name(name string) {
	p.name = name
}

func (p *pluginBuilder) Description(description string) {
	p.description = description
}

func (p *pluginBuilder) Version(v int) {
	p.version = v
}

func (p *pluginBuilder) Dependencies(dependentPluginWithVersion map[string]int) {
	p.dependentPluginWithVersion = dependentPluginWithVersion
}

func (p *pluginBuilder) OnLoad(f func(registration lsha.ModeRepository)) {
	p.onLoad = f
}
