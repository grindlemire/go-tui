package tui

import "sync"

// Events is a simple event bus for cross-component communication.
// It is generic over the event type T.
type Events[T any] struct {
	mu        sync.RWMutex
	listeners []func(T)
	app       *App
}

// NewEvents creates a new event bus.
// The bus is created unbound — it will be bound to an App later via
// BindApp (called by generated code during mount).
func NewEvents[T any]() *Events[T] {
	return &Events[T]{}
}

// NewEventsForApp creates an event bus bound to the provided app.
func NewEventsForApp[T any](app *App) *Events[T] {
	if app == nil {
		panic("tui: nil app in NewEventsForApp")
	}
	return &Events[T]{app: app}
}

// BindApp binds this event bus to the given app for dirty-marking.
// Panics if app is nil. Idempotent for the same app; overwrites if different.
func (e *Events[T]) BindApp(app *App) {
	if app == nil {
		panic("tui: nil app in Events.BindApp")
	}
	e.mu.Lock()
	e.app = app
	e.mu.Unlock()
}

// Emit sends an event to all listeners and marks the UI dirty.
func (e *Events[T]) Emit(event T) {
	app := e.resolveApp()
	e.mu.RLock()
	listeners := e.listeners
	e.mu.RUnlock()

	for _, fn := range listeners {
		fn(event)
	}
	app.MarkDirty()
}

// Subscribe adds a listener for events.
func (e *Events[T]) Subscribe(fn func(T)) {
	e.mu.Lock()
	e.listeners = append(e.listeners, fn)
	e.mu.Unlock()
}

func (e *Events[T]) resolveApp() *App {
	e.mu.RLock()
	app := e.app
	e.mu.RUnlock()
	if app != nil {
		return app
	}
	panic("tui.Events used without app context; call BindApp or use NewEventsForApp")
}
