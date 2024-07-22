package core

import (
	"github.com/ohanan/LambdaSha/pkg/lsha"
)

type TriggerBuilder struct {
	priority float64
	invoker  func(event lsha.Event, result lsha.InvokeResult)
}

func (t *TriggerBuilder) Priority(priority float64) lsha.TriggerBuilder {
	t.priority = priority
	return t
}

func (t *TriggerBuilder) OnInvoke(f func(event lsha.Event, result lsha.InvokeResult)) lsha.TriggerBuilder {
	t.invoker = f
	return t
}

type InvokerResult struct {
}

func (i *InvokerResult) FastStop() {
}

type Trigger struct {
	id        uint64
	name      string
	eventName string
	priority  float64
	invoker   lsha.Invoker
}

func (t *Trigger) ID() uint64 {
	return t.id
}

func (t *Trigger) Name() string {
	return t.name
}

func (t *Trigger) EventName() string {
	return t.eventName
}

func (t *Trigger) Priority() float64 {
	return t.priority
}
