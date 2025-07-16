package events

import (
	"maps"
	"sync"
	"sync/atomic"
)

type Emitter[T any] struct {
	id        int64
	listeners map[int64]func(T)
	lock      sync.RWMutex
}

func (e *Emitter[T]) Events() Event[T] {
	return e
}

func (e *Emitter[T]) Subscribe(listener func(T)) int64 {
	id := atomic.AddInt64(&e.id, 1)
	e.lock.Lock()
	if e.listeners == nil {
		e.listeners = make(map[int64]func(T))
	}
	e.listeners[id] = listener
	e.lock.Unlock()
	return id
}

func (e *Emitter[T]) Unsubscribe(id int64) {
	e.lock.Lock()
	defer e.lock.Unlock()
	delete(e.listeners, id)
}

func (e *Emitter[T]) Emit(data T) {
	e.lock.RLock()
	cp := maps.Clone(e.listeners)
	e.lock.RUnlock()
	for _, listener := range cp {
		listener(data)
	}
}

type Event[T any] interface {
	Subscribe(listener func(T)) int64
	Unsubscribe(id int64)
}
