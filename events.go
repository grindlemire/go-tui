package tui

import (
	"fmt"
	"reflect"
	"sync"
)

// Events is a simple event bus for cross-component communication.
// It is generic over the event type T and routes by topic key.
type Events[T any] struct {
	mu          sync.RWMutex
	subscribers []*eventSubscriber[T]
	app         *App
	topic       string
	eventType   reflect.Type
}

type eventSubscriber[T any] struct {
	fn          func(T)
	active      bool
	unsubscribe func()
}

type topicSubscription struct {
	eventType reflect.Type
	listeners map[uint64]func(any)
	nextID    uint64
}

// NewEvents creates a new topic-based event bus.
// The bus is created unbound — it will be bound to an App later via
// BindApp (called by generated code during mount).
func NewEvents[T any](topic string) *Events[T] {
	if topic == "" {
		panic("tui: empty topic in NewEvents")
	}
	return &Events[T]{topic: topic, eventType: eventTypeOf[T]()}
}

// NewEventsForApp creates an event bus bound to the provided app.
func NewEventsForApp[T any](app *App, topic string) *Events[T] {
	if app == nil {
		panic("tui: nil app in NewEventsForApp")
	}
	ev := NewEvents[T](topic)
	ev.BindApp(app)
	return ev
}

// BindApp binds this event bus to the given app for dirty-marking.
// Panics if app is nil. Idempotent for the same app; overwrites if different.
func (e *Events[T]) BindApp(app *App) {
	if app == nil {
		panic("tui: nil app in Events.BindApp")
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.app == app {
		e.bindSubscribersLocked(app)
		return
	}
	e.unbindSubscribersLocked()
	e.app = app
	e.bindSubscribersLocked(app)
}

// UnbindApp removes this bus from app topic routing.
func (e *Events[T]) UnbindApp() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.unbindSubscribersLocked()
	e.app = nil
}

// Emit sends an event to all listeners and marks the UI dirty.
func (e *Events[T]) Emit(event T) {
	app := e.resolveApp()
	app.publishTopic(e.topic, e.eventType, event)
	app.MarkDirty()
}

// Subscribe adds a listener for events.
func (e *Events[T]) Subscribe(fn func(T)) func() {
	if fn == nil {
		return func() {}
	}

	e.mu.Lock()
	sub := &eventSubscriber[T]{fn: fn, active: true}
	e.subscribers = append(e.subscribers, sub)
	if e.app != nil {
		unsub, err := e.app.subscribeTopic(e.topic, e.eventType, func(v any) {
			sub.fn(v.(T))
		})
		if err == nil {
			sub.unsubscribe = unsub
		}
	}
	e.mu.Unlock()

	return func() {
		e.mu.Lock()
		defer e.mu.Unlock()
		if !sub.active {
			return
		}
		sub.active = false
		if sub.unsubscribe != nil {
			sub.unsubscribe()
			sub.unsubscribe = nil
		}
	}
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

func (e *Events[T]) bindSubscribersLocked(app *App) {
	for _, sub := range e.subscribers {
		if !sub.active || sub.unsubscribe != nil {
			continue
		}
		unsub, err := app.subscribeTopic(e.topic, e.eventType, func(v any) {
			sub.fn(v.(T))
		})
		if err == nil {
			sub.unsubscribe = unsub
		}
	}
}

func (e *Events[T]) unbindSubscribersLocked() {
	for _, sub := range e.subscribers {
		if sub.unsubscribe != nil {
			sub.unsubscribe()
			sub.unsubscribe = nil
		}
	}
}

func (a *App) subscribeTopic(topic string, eventType reflect.Type, fn func(any)) (func(), error) {
	if a == nil {
		panic("tui: nil app in subscribeTopic")
	}
	if topic == "" {
		panic("tui: empty topic in subscribeTopic")
	}
	if fn == nil {
		return func() {}, nil
	}

	a.topicMu.Lock()
	if a.topics == nil {
		a.topics = make(map[string]*topicSubscription)
	}

	sub, exists := a.topics[topic]
	if !exists {
		sub = &topicSubscription{
			eventType: eventType,
			listeners: make(map[uint64]func(any)),
		}
		a.topics[topic] = sub
	} else if sub.eventType != eventType {
		a.topicMu.Unlock()
		return nil, fmt.Errorf("tui: topic type mismatch for %s: existing=%s, new=%s", topic, sub.eventType, eventType)
	}

	id := sub.nextID
	sub.nextID++
	sub.listeners[id] = fn
	a.topicMu.Unlock()

	removed := false
	return func() {
		a.topicMu.Lock()
		defer a.topicMu.Unlock()
		if removed {
			return
		}
		removed = true
		if sub, ok := a.topics[topic]; ok {
			delete(sub.listeners, id)
		}
	}, nil
}

func (a *App) publishTopic(topic string, eventType reflect.Type, event any) {
	if a == nil {
		panic("tui: nil app in publishTopic")
	}
	if topic == "" {
		panic("tui: empty topic in publishTopic")
	}

	a.topicMu.RLock()
	sub, ok := a.topics[topic]
	if !ok {
		a.topicMu.RUnlock()
		return
	}
	if sub.eventType != eventType {
		a.topicMu.RUnlock()
		return
	}

	listeners := make([]func(any), 0, len(sub.listeners))
	for _, fn := range sub.listeners {
		listeners = append(listeners, fn)
	}
	a.topicMu.RUnlock()

	for _, fn := range listeners {
		fn(event)
	}
}

func eventTypeOf[T any]() reflect.Type {
	var t *T
	return reflect.TypeOf(t).Elem()
}
