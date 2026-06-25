package tui

import (
	"fmt"

	"github.com/grindlemire/go-tui/internal/debug"
)

// focusQuerier is implemented by components that can report their own focus state.
// Used by the dispatch table to evaluate focus-gated key bindings.
type focusQuerier interface {
	IsFocused() bool
}

// dispatchEntry is a handler with its tree position for ordering.
type dispatchEntry struct {
	pattern    KeyPattern
	handler    func(KeyEvent)
	stop       bool
	preempt    bool        // fires before normal handlers (used by modal)
	position   int         // BFS order index from tree walk
	focusCheck func() bool // Non-nil for focus-gated entries; returns true when component is focused
}

// dispatchTable holds all handlers in a single tree-ordered list.
// Handlers are matched against incoming KeyEvents by pattern.
type dispatchTable struct {
	entries []dispatchEntry // All handlers, ordered by tree position
}

// buildDispatchTable walks the element tree, collects KeyMap() from
// all mounted components, validates exclusive conflicts, and builds
// the dispatch table ordered by tree position.
func buildDispatchTable(rootComp Component, root *Element) (*dispatchTable, error) {
	table := &dispatchTable{}
	position := 0

	walkComponents(rootComp, root, func(comp Component) {
		kl, ok := comp.(KeyListener)
		if !ok {
			return
		}
		km := kl.KeyMap()
		if km == nil {
			return
		}

		fq, hasFocusQuery := comp.(focusQuerier)

		for _, binding := range km {
			entry := dispatchEntry{
				pattern:  binding.Pattern,
				handler:  binding.Handler,
				stop:     binding.Stop,
				preempt:  binding.Preempt,
				position: position,
			}
			// For focus-gated bindings, capture the component's focus check
			if binding.Pattern.FocusRequired && hasFocusQuery {
				entry.focusCheck = fq.IsFocused
			}
			table.entries = append(table.entries, entry)
		}
		position++
	})

	if err := table.validate(); err != nil {
		return nil, err
	}

	return table, nil
}

// matchesKey checks if a dispatch entry's key pattern matches a key event,
// without checking focus state.
func (e *dispatchEntry) matchesKey(ke KeyEvent) bool {
	p := e.pattern

	if p.AnyKey {
		return true
	}

	if p.ExcludeMods != 0 && ke.Mod&p.ExcludeMods != 0 {
		return false
	}
	if p.Mod != 0 && ke.Mod != p.Mod {
		return false
	}

	if p.AnyRune && ke.Key == KeyRune {
		return true
	}
	if p.Rune != 0 && ke.Rune == p.Rune && ke.Key == KeyRune {
		return true
	}
	if p.Key != 0 && ke.Key == p.Key {
		return true
	}
	return false
}

// matches checks if a dispatch entry matches a key event, including focus gating.
func (e *dispatchEntry) matches(ke KeyEvent) bool {
	if e.pattern.FocusRequired && e.focusCheck != nil {
		if !e.focusCheck() {
			return false
		}
	}
	return e.matchesKey(ke)
}

// dispatch sends a key event to matching handlers.
// Focus-gated stop handlers take priority: if any active focus-gated stop
// handler matches, it fires exclusively and broadcast handlers are skipped.
// Otherwise, handlers fire in tree order, stopping early if a Stop handler matches.
func (dt *dispatchTable) dispatch(ke KeyEvent) bool {
	if dt == nil {
		return false
	}
	debug.Topic("dispatch", "Key=%s Rune=%q Mod=%s (entries=%d)", ke.Key, ke.Rune, ke.Mod, len(dt.entries))

	// Priority pass: focus-gated stop handlers consume the event exclusively.
	// This ensures a focused input captures keys like 'q' before broadcast
	// handlers (like quit) can intercept them.
	for i := range dt.entries {
		e := &dt.entries[i]
		if e.pattern.FocusRequired && e.stop && e.focusCheck != nil && e.focusCheck() {
			if e.matchesKey(ke) {
				e.handler(ke)
				return true
			}
		}
	}

	// Preemptive pass: overlay handlers that must fire before normal dispatch.
	// Used by modal to block parent handlers from seeing key events.
	for i := range dt.entries {
		if !dt.entries[i].preempt {
			continue
		}
		if dt.entries[i].matches(ke) {
			dt.entries[i].handler(ke)
			if dt.entries[i].stop {
				return true
			}
		}
	}

	// Normal dispatch: broadcast and non-stop handlers in tree order.
	for i := range dt.entries {
		if dt.entries[i].preempt {
			continue // already handled
		}
		if dt.entries[i].matches(ke) {
			debug.Topic("dispatch", "matched entry[%d] pattern={Key=%s Rune=%q Mod=%v ExcludeMods=%v} stop=%v",
				i, dt.entries[i].pattern.Key, dt.entries[i].pattern.Rune, dt.entries[i].pattern.Mod, dt.entries[i].pattern.ExcludeMods, dt.entries[i].stop)
			dt.entries[i].handler(ke)
			if dt.entries[i].stop {
				return true
			}
		}
	}
	return false
}

// validate checks for conflicting Stop handlers. Two active Stop handlers
// for the same key pattern is an error — it's ambiguous which should win.
// A Stop handler + a broadcast handler for the same pattern is fine.
func (dt *dispatchTable) validate() error {
	// Track patterns that already have a Stop handler
	type stopInfo struct {
		position    int
		droppedGate bool
	}
	stopPatterns := make(map[KeyPattern]stopInfo)

	for _, entry := range dt.entries {
		if !entry.stop {
			continue
		}
		// A focus-gated entry is exempt only when its focus check resolved: the
		// owning component implements IsFocused (and was mounted), so at most one
		// such binding is active at a time. When focusCheck is nil the gate was
		// dropped — the binding fires unconditionally — so treat it as an ordinary
		// stop handler that can conflict.
		if entry.pattern.FocusRequired && entry.focusCheck != nil {
			continue
		}
		// Preemptive entries (modal overlay) run in a separate dispatch pass
		// and cannot conflict with normal stop handlers.
		if entry.preempt {
			continue
		}
		// Reaching here with FocusRequired means focusCheck was nil: the focus gate
		// was dropped (owner has no IsFocused, or the widget was not mounted).
		droppedGate := entry.pattern.FocusRequired
		// Strip FocusRequired for comparison so focus-gated and broadcast entries
		// with the same key don't conflict
		comparePattern := entry.pattern
		comparePattern.FocusRequired = false
		if existing, conflict := stopPatterns[comparePattern]; conflict {
			if droppedGate || existing.droppedGate {
				return fmt.Errorf(
					"focus-gated key binding for pattern %+v at tree positions %d and %d fires "+
						"unconditionally because its owning component does not implement IsFocused() "+
						"or was not mounted; mount each focusable widget as its own component "+
						"(app.Mount) instead of aggregating their KeyMaps onto a host",
					entry.pattern, existing.position, entry.position,
				)
			}
			return fmt.Errorf(
				"conflicting stop handlers for key pattern %+v at tree positions %d and %d",
				entry.pattern, existing.position, entry.position,
			)
		}
		stopPatterns[comparePattern] = stopInfo{position: entry.position, droppedGate: droppedGate}
	}

	return nil
}
