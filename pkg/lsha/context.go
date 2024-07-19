package lsha

type Context interface {
	RuntimeContext
	WithValue(key string, value any) Context
	WithValues(values map[string]any) Context
	GetValue(key string) (value any, ok bool)
	VisitAllValues(key string, callback func(v any) (willContinue bool))
}
type RuntimeContext interface {
	GetUsers() []User
	GetModeConfig() ModeConfig
	GetMode() Mode
	GetCurrentTurn() Turn
}

func Value[V any](c Context, key string) (value V, ok bool) {
	v, ok := c.GetValue(key)
	if !ok {
		return
	}
	value, ok = v.(V)
	return
}
func Visit[V any](c Context, key string, callback func(v V) (willContinue bool)) {
	c.VisitAllValues(key, func(v any) bool {
		if vv, ok := v.(V); ok {
			return callback(vv)
		}
		return true
	})
}
func VisitAll[V any](c Context, key string, callback func(v V)) {
	c.VisitAllValues(key, func(v any) bool {
		if vv, ok := v.(V); ok {
			callback(vv)
		}
		return true
	})
}
