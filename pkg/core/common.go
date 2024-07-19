package core

import "sync/atomic"

func atomicAppend[T any](a *atomic.Pointer[[]T], values ...T) {
	load := *a.Load()
	load = append(load, values...)
	a.Store(&load)
}

func firstNotNil[T any](a []*T) (int, *T) {
	for idx, v := range a {
		if v != nil {
			return idx, v
		}
	}
	return -1, nil
}
