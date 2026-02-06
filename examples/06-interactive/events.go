package main

import (
	"sync"

	tui "github.com/grindlemire/go-tui"
)

// Events is a simple event bus for cross-component communication.
type Events[T any] struct {
	mu        sync.RWMutex
	listeners []func(T)
}

// NewEvents creates a new event bus.
func NewEvents[T any]() *Events[T] {
	return &Events[T]{}
}

// Emit sends an event to all listeners.
func (e *Events[T]) Emit(event T) {
	e.mu.RLock()
	listeners := e.listeners
	e.mu.RUnlock()

	for _, fn := range listeners {
		fn(event)
	}
	tui.MarkDirty()
}

// Subscribe adds a listener for events.
func (e *Events[T]) Subscribe(fn func(T)) {
	e.mu.Lock()
	e.listeners = append(e.listeners, fn)
	e.mu.Unlock()
}
