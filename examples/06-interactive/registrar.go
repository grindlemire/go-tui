package main

import (
	"time"

	tui "github.com/grindlemire/go-tui"
)

// Registrar collects event bindings during OnMount.
// This is a prototype of what could be a framework-level API.
type Registrar struct {
	keys    tui.KeyMap
	clicks  []clickBinding
	timers  []timerBinding
}

type clickBinding struct {
	ref *tui.Ref
	fn  func()
}

type timerBinding struct {
	interval time.Duration
	fn       func()
}

// NewRegistrar creates a new registrar for collecting bindings.
func NewRegistrar() *Registrar {
	return &Registrar{}
}

// --- Key bindings ---

// OnRune registers a handler for a printable character.
func (r *Registrar) OnRune(ru rune, fn func()) {
	r.keys = append(r.keys, tui.OnRune(ru, func(ke tui.KeyEvent) { fn() }))
}

// OnKey registers a handler for a special key.
func (r *Registrar) OnKey(key tui.Key, fn func()) {
	r.keys = append(r.keys, tui.OnKey(key, func(ke tui.KeyEvent) { fn() }))
}

// OnKeyStop registers a handler that stops event propagation.
func (r *Registrar) OnKeyStop(key tui.Key, fn func()) {
	r.keys = append(r.keys, tui.OnKeyStop(key, func(ke tui.KeyEvent) { fn() }))
}

// --- Mouse bindings ---

// OnClick registers a click handler for a ref. The framework handles hit testing.
func (r *Registrar) OnClick(ref *tui.Ref, fn func()) {
	r.clicks = append(r.clicks, clickBinding{ref: ref, fn: fn})
}

// --- Timer bindings ---

// Every registers a function to be called at a regular interval.
func (r *Registrar) Every(interval time.Duration, fn func()) {
	r.timers = append(r.timers, timerBinding{interval: interval, fn: fn})
}

// --- Accessors for the app to use ---

// KeyMap returns the collected key bindings.
func (r *Registrar) KeyMap() tui.KeyMap {
	return r.keys
}

// HandleMouse checks clicks against registered refs and calls handlers.
func (r *Registrar) HandleMouse(me tui.MouseEvent) bool {
	if me.Button == tui.MouseLeft && me.Action == tui.MousePress {
		for _, cb := range r.clicks {
			if cb.ref.El() != nil && cb.ref.El().ContainsPoint(me.X, me.Y) {
				cb.fn()
				return true
			}
		}
	}
	return false
}

// Timers returns the collected timer bindings as tui.OnTimer calls.
// These need to be attached to an element's onTimer attribute.
func (r *Registrar) Timers() []timerBinding {
	return r.timers
}

// HasTimers returns true if any timers were registered.
func (r *Registrar) HasTimers() bool {
	return len(r.timers) > 0
}

// --- Merge multiple registrars ---

// Merge combines another registrar's bindings into this one.
func (r *Registrar) Merge(other *Registrar) {
	r.keys = append(r.keys, other.keys...)
	r.clicks = append(r.clicks, other.clicks...)
	r.timers = append(r.timers, other.timers...)
}
