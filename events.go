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
func NewEvents[T any]() *Events[T] {
	app := DefaultApp()
	if app == nil {
		panic("tui.NewEvents requires a default app; call SetDefaultApp or use NewEventsForApp")
	}
	return NewEventsForApp[T](app)
}

// NewEventsForApp creates an event bus bound to the provided app.
func NewEventsForApp[T any](app *App) *Events[T] {
	if app == nil {
		panic("tui: nil app in NewEvents")
	}
	return &Events[T]{app: app}
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
	app = DefaultApp()
	if app == nil {
		panic("tui.Events used without app context; use NewEventsForApp or SetDefaultApp")
	}
	e.mu.Lock()
	if e.app == nil {
		e.app = app
	}
	e.mu.Unlock()
	return app
}
