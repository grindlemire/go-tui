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
//
// Batching:
//
// Use Batch() to coalesce multiple Set() calls and avoid redundant binding
// execution:
//
//	tui.Batch(func() {
//	    firstName.Set("Bob")
//	    lastName.Set("Smith")
//	})  // Bindings fire once here, not twice
package tui

import (
	"sync"
	"sync/atomic"

	"github.com/grindlemire/go-tui/pkg/debug"
)

// batchContext tracks batch state for deferring binding execution.
type batchContext struct {
	mu           sync.Mutex
	depth        int               // nesting depth (0 = not batching)
	pending      map[uint64]func() // pending binding callbacks keyed by binding ID
	pendingOrder []uint64          // order in which bindings were first triggered
}

// batchCtx is the package-level batch context.
var batchCtx = batchContext{
	pending: make(map[uint64]func()),
}

// globalBindingID is a global counter for generating unique binding IDs.
// This ensures binding IDs are unique across all State instances.
var globalBindingID atomic.Uint64

// State wraps a value and notifies bindings when it changes.
// State is generic over any type T.
type State[T any] struct {
	mu       sync.RWMutex
	value    T
	bindings []*binding[T]
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
// If called within a Batch(), binding execution is deferred until the
// batch completes.
func (s *State[T]) Set(v T) {
	debug.Log("State.Set: setting value to %v", v)
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

	// Check if we're in a batch
	batchCtx.mu.Lock()
	isBatching := batchCtx.depth > 0
	if isBatching {
		// Defer binding execution - store closures keyed by binding ID
		// Later Set() calls to same binding ID will overwrite with new value
		// Track order of first occurrence for deterministic execution order
		for _, b := range activeBindings {
			bindingID := b.id
			bindingFn := b.fn
			capturedValue := v
			if _, exists := batchCtx.pending[bindingID]; !exists {
				// First time seeing this binding, track its order
				batchCtx.pendingOrder = append(batchCtx.pendingOrder, bindingID)
			}
			batchCtx.pending[bindingID] = func() { bindingFn(capturedValue) }
		}
	}
	batchCtx.mu.Unlock()

	// Execute bindings immediately if not batching
	if !isBatching {
		debug.Log("State.Set: executing %d bindings immediately", len(activeBindings))
		for _, b := range activeBindings {
			b.fn(v)
		}
	} else {
		debug.Log("State.Set: deferred %d bindings (batching)", len(activeBindings))
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
	// Use global counter to ensure unique IDs across all State instances
	id := globalBindingID.Add(1)

	s.mu.Lock()
	b := &binding[T]{id: id, fn: fn, active: true}
	s.bindings = append(s.bindings, b)
	s.mu.Unlock()

	return func() {
		s.mu.Lock()
		b.active = false
		s.mu.Unlock()
	}
}

// Batch executes fn and defers all binding callbacks until fn returns.
// Use this when updating multiple states to avoid redundant element updates.
//
// When the same binding is triggered multiple times during a batch,
// it only executes once with the final value.
//
// Bindings are executed in the order they were first triggered during the batch.
// This provides deterministic ordering for bindings across different states.
//
// Nested Batch calls are supported - bindings only fire when the outermost
// Batch completes.
//
// If fn panics, the batch state is properly cleaned up before the panic
// propagates.
//
// Example:
//
//	tui.Batch(func() {
//	    firstName.Set("Bob")
//	    lastName.Set("Smith")
//	    age.Set(30)
//	})
//	// Bindings fire once here, not three times
func Batch(fn func()) {
	batchCtx.mu.Lock()
	batchCtx.depth++
	batchCtx.mu.Unlock()

	defer func() {
		batchCtx.mu.Lock()
		batchCtx.depth--
		shouldExecute := batchCtx.depth == 0 && len(batchCtx.pending) > 0
		var pendingCallbacks []func()
		if shouldExecute {
			// Collect callbacks in the order they were first triggered
			pendingCallbacks = make([]func(), 0, len(batchCtx.pendingOrder))
			for _, id := range batchCtx.pendingOrder {
				if callback, exists := batchCtx.pending[id]; exists {
					pendingCallbacks = append(pendingCallbacks, callback)
				}
			}
			batchCtx.pending = make(map[uint64]func())
			batchCtx.pendingOrder = nil
		}
		batchCtx.mu.Unlock()

		// Execute callbacks outside the lock
		if shouldExecute {
			for _, callback := range pendingCallbacks {
				callback()
			}
		}
	}()

	fn()
}

// TestResetBatch resets the batch context state for testing.
// Only use this in test code.
func TestResetBatch() {
	batchCtx.mu.Lock()
	batchCtx.depth = 0
	batchCtx.pending = make(map[uint64]func())
	batchCtx.pendingOrder = nil
	batchCtx.mu.Unlock()
}
