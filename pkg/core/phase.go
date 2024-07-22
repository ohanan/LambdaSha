package core

import (
	"github.com/ohanan/LambdaSha/pkg/lsha"
)

type Phase struct {
	data any
	name string
}

func (p *Phase) BindData(data any) {
	p.data = data
}

func (p *Phase) Data() any {
	return p.data
}

func (p *Phase) Name() string {
	return p.name
}

type PhaseBuilder struct {
	name string
}

func (p *PhaseBuilder) Name(name string) lsha.PhaseBuilder {
	p.name = name
	return p
}
