package xsync

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
)

type Task[T any] struct {
	state  T
	cancel func()
	done   chan struct{}
}

func (t *Task[T]) State() T {
	return t.state
}

func (t *Task[T]) Wait() {
	<-t.done
}

func (t *Task[T]) GetIfRunning() (out T) {
	select {
	case <-t.done:
		return t.state
	default:
		return
	}
}

func (t *Task[T]) Stop(ctx context.Context) {
	t.cancel()
	select {
	case <-ctx.Done():
	case <-t.done:
	}

}

func (t *Task[T]) Done() <-chan struct{} {
	return t.done
}

func (t *Task[T]) IsRunning() bool {
	select {
	case <-t.done:
		return false
	default:
		return true
	}
}

func Spawn[T any](ctx context.Context, state T, fn func(ctx context.Context)) *Task[T] {
	done := make(chan struct{})
	child, cancel := context.WithCancel(ctx)

	t := &Task[T]{
		state:  state,
		cancel: cancel,
		done:   done,
	}

	go func() {
		defer close(done)
		fn(child)
	}()

	return t
}

func NewPool[T any]() *Pool[T] {
	return &Pool[T]{
		fanout: make(chan func(ctx context.Context) *Task[T]),
	}
}

type Pool[T any] struct {
	sequence   atomic.Int64
	inProgress sync.Map // int64 -> *Task
	fanout     chan func(ctx context.Context) *Task[T]
}

func (p *Pool[T]) Try(state T, fn func(ctx context.Context)) error {
	task := func(ctx context.Context) *Task[T] {
		return Spawn(ctx, state, fn)
	}

	select {
	case p.fanout <- task:
		return nil
	default:
		return errors.New("pool is full")
	}
}

func (p *Pool[T]) Run(ctx context.Context) {
	for {
		select {
		case task := <-p.fanout:
			t := task(ctx)

			h := p.addTask(t)
			<-t.Done()
			p.deleteTask(h)

		case <-ctx.Done():
			return
		}
	}
}

func (p *Pool[T]) List() []*Task[T] {
	var tasks []*Task[T]
	p.inProgress.Range(func(key, value any) bool {
		tasks = append(tasks, value.(*Task[T]))
		return true
	})
	return tasks
}

func (p *Pool[T]) RunningStates() []T {
	var tasks []T
	p.inProgress.Range(func(key, value any) bool {
		if t, ok := value.(*Task[T]); ok && t.IsRunning() {
			tasks = append(tasks, t.State())
		}
		return true
	})
	return tasks
}

func (p *Pool[T]) addTask(t *Task[T]) int64 {
	id := p.sequence.Add(1)
	p.inProgress.Store(id, t)
	return id
}

func (p *Pool[T]) deleteTask(id int64) {
	p.inProgress.Delete(id)
}
