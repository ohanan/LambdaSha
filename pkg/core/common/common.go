package common

import "sync/atomic"

func AtomicAppend[T any](a *atomic.Pointer[[]T], values ...T) {
	load := *a.Load()
	load = append(load, values...)
	a.Store(&load)
}

func FirstNotNil[T any](a []*T) (int, *T) {
	for idx, v := range a {
		if v != nil {
			return idx, v
		}
	}
	return -1, nil
}

func Ptr[V any](v V) *V {
	return &v
}

func SliceToStructMap[K comparable](a []K) map[K]struct{} {
	m := make(map[K]struct{}, len(a))
	for _, v := range a {
		m[v] = struct{}{}
	}
	return m
}
