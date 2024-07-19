package core

import (
	"github.com/ohanan/LambdaSha/pkg/lsha"
)

var _ lsha.RuntimeContext = (*runtimeContext)(nil)

func newContext() *Context {
	return &Context{
		runtimeContext: &runtimeContext{},
		parent:         nil,
		values:         map[string]any{},
	}
}

type runtimeContext struct {
	config      lsha.ModeConfig
	mode        lsha.Mode
	currentTurn lsha.Turn
	accounts    []lsha.User
}

func (c *runtimeContext) GetUsers() []lsha.User {
	return c.accounts
}

func (c *runtimeContext) GetModeConfig() lsha.ModeConfig {
	return c.config
}

func (c *runtimeContext) GetMode() lsha.Mode {
	return c.mode
}

func (c *runtimeContext) GetCurrentTurn() lsha.Turn {
	return c.currentTurn
}

type Context struct {
	*runtimeContext
	parent *Context
	values map[string]any
}

func (c *Context) WithValue(key string, value any) lsha.Context {
	return &Context{
		runtimeContext: c.runtimeContext,
		parent:         c,
		values:         map[string]any{key: value},
	}
}
func (c *Context) WithValues(values map[string]any) lsha.Context {
	c2 := &Context{
		runtimeContext: c.runtimeContext,
		parent:         c,
		values:         map[string]any{},
	}
	for k, v := range values {
		c2.values[k] = v
	}
	return c2
}

func (c *Context) GetValue(key string) (value any, ok bool) {
	cc := c
	for cc != nil {
		if value, ok = cc.values[key]; ok {
			return
		}
		cc = cc.parent
	}
	return nil, false
}

func (c *Context) VisitAllValues(key string, callback func(v any) bool) {
	for cc := c; cc != nil; cc = cc.parent {
		if value, ok := cc.values[key]; ok {
			callback(value)
		}
	}
}
