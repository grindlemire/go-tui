// Package tui provides the core State type for reactive UI bindings.
//
// State[T] wraps a value and notifies bindings when it changes. This enables
// automatic UI updates without manual SetText() calls.
//
// Thread Safety Rules:
//   - Get() is safe to call from any goroutine
//   - Set() must only be called from the main event loop
//   - For background updates, use channel watchers or App.QueueUpdate()
//
// Example usage:
//
//	count := tui.NewState(0)
//	count.Bind(func(v int) {
//	    span.SetText(fmt.Sprintf("Count: %d", v))
//	})
//	count.Set(count.Get() + 1)  // triggers binding and marks dirty
package tui

import "sync"

// State wraps a value and notifies bindings when it changes.
// State is generic over any type T.
type State[T any] struct {
	mu       sync.RWMutex
	value    T
	bindings []*binding[T]
	nextID   uint64
}

// binding represents a registered callback that fires when state changes.
type binding[T any] struct {
	id     uint64
	fn     func(T)
	active bool
}

// Unbind is a handle to remove a binding. Call it to prevent
// future callback invocations for the associated binding.
type Unbind func()

// NewState creates a new state with the given initial value.
// The type T is inferred from the initial value.
//
// Example:
//
//	count := tui.NewState(0)           // State[int]
//	name := tui.NewState("hello")      // State[string]
//	items := tui.NewState([]string{})  // State[[]string]
func NewState[T any](initial T) *State[T] {
	return &State[T]{value: initial}
}

// Get returns the current value. Thread-safe for reading from any goroutine.
func (s *State[T]) Get() T {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.value
}

// Set updates the value, marks dirty, and notifies all bindings.
//
// IMPORTANT: Must be called from main loop only. For background
// updates, use app.QueueUpdate() or channel watchers.
//
// Set automatically calls MarkDirty() to trigger a re-render.
func (s *State[T]) Set(v T) {
	s.mu.Lock()
	s.value = v
	// Copy active bindings while holding lock and remove inactive ones
	// to prevent memory leaks from accumulated unbound bindings
	activeBindings := make([]*binding[T], 0, len(s.bindings))
	for _, b := range s.bindings {
		if b.active {
			activeBindings = append(activeBindings, b)
		}
	}
	// Replace bindings slice with only active bindings (cleanup)
	s.bindings = activeBindings
	s.mu.Unlock()

	// Mark dirty using existing atomic flag
	MarkDirty()

	// Execute bindings outside lock (they may call Get())
	// Note: Batching support will be added in Phase 2
	for _, b := range activeBindings {
		b.fn(v)
	}
}

// Update applies a function to the current value and sets the result.
// This is a convenience method for read-modify-write operations.
//
// Example:
//
//	count.Update(func(v int) int { return v + 1 })
func (s *State[T]) Update(fn func(T) T) {
	s.Set(fn(s.Get()))
}

// Bind registers a function to be called when the value changes.
// Returns an Unbind handle to remove the binding.
//
// The binding callback receives the new value as its argument.
// Bindings are executed in registration order.
//
// Example:
//
//	unbind := count.Bind(func(v int) {
//	    fmt.Println("count changed to", v)
//	})
//	// Later, to stop receiving updates:
//	unbind()
func (s *State[T]) Bind(fn func(T)) Unbind {
	s.mu.Lock()
	id := s.nextID
	s.nextID++
	b := &binding[T]{id: id, fn: fn, active: true}
	s.bindings = append(s.bindings, b)
	s.mu.Unlock()

	return func() {
		s.mu.Lock()
		b.active = false
		s.mu.Unlock()
	}
}
