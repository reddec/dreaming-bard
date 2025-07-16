package xsync

import (
	"sync"
)

type Value[T any] struct {
	value T
	lock  sync.RWMutex
}

func (v *Value[T]) Load() T {
	v.lock.RLock()
	defer v.lock.RUnlock()
	return v.value
}

func (v *Value[T]) Store(value T) {
	v.lock.Lock()
	defer v.lock.Unlock()
	v.value = value
}

func (v *Value[T]) Swap(value T) T {
	v.lock.Lock()
	defer v.lock.Unlock()
	old := v.value
	v.value = value
	return old
}
