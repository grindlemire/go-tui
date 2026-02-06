package main

import "sync"

// Events is a typed event bus for fire-and-forget event emission.
// Unlike State[T], it doesn't store a value - it just notifies subscribers.
type Events[T any] struct {
	mu        sync.RWMutex
	listeners []func(T)
}

// NewEvents creates a new typed event bus.
func NewEvents[T any]() *Events[T] {
	return &Events[T]{}
}

// Emit sends an event to all subscribers.
func (e *Events[T]) Emit(value T) {
	e.mu.RLock()
	listeners := e.listeners
	e.mu.RUnlock()

	for _, fn := range listeners {
		fn(value)
	}
}

// Subscribe registers a callback to be called on each emitted event.
func (e *Events[T]) Subscribe(fn func(T)) {
	e.mu.Lock()
	e.listeners = append(e.listeners, fn)
	e.mu.Unlock()
}
