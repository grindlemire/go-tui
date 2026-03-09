package tui

import (
	"fmt"

	"github.com/grindlemire/go-tui/internal/debug"
)

// dispatchEntry is a handler with its tree position for ordering.
type dispatchEntry struct {
	pattern    KeyPattern
	handler    func(KeyEvent)
	stop       bool
	position   int        // BFS order index from tree walk
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
func buildDispatchTable(rootComp Component, root *Element, fm *focusManager) (*dispatchTable, error) {
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

		debug.Log("buildDispatchTable: component %T at position %d has %d bindings", comp, position, len(km))

		// Check if this component can report its own focus state
		type focusQuerier interface {
			IsFocused() bool
		}
		fq, hasFocusQuery := comp.(focusQuerier)
		if hasFocusQuery {
			debug.Log("buildDispatchTable:   component %T implements focusQuerier, IsFocused=%v", comp, fq.IsFocused())
		}

		for _, binding := range km {
			entry := dispatchEntry{
				pattern:  binding.Pattern,
				handler:  binding.Handler,
				stop:     binding.Stop,
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

	if p.RequireNoMods && ke.Mod != 0 {
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
func (e *dispatchEntry) matches(ke KeyEvent, fm *focusManager) bool {
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
func (dt *dispatchTable) dispatch(ke KeyEvent, fm *focusManager) bool {
	if dt == nil {
		return false
	}

	debug.Log("dispatchTable.dispatch: key=%v rune=%c mod=%v (entries=%d)", ke.Key, ke.Rune, ke.Mod, len(dt.entries))

	// Priority pass: focus-gated stop handlers consume the event exclusively.
	// This ensures a focused input captures keys like 'q' before broadcast
	// handlers (like quit) can intercept them.
	for i := range dt.entries {
		e := &dt.entries[i]
		if e.pattern.FocusRequired && e.stop && e.focusCheck != nil && e.focusCheck() {
			if e.matchesKey(ke) {
				debug.Log("dispatchTable.dispatch: PRIORITY focus-gated stop handler fired at position %d, pattern=%+v", e.position, e.pattern)
				e.handler(ke)
				return true
			}
		}
	}

	// Normal dispatch: broadcast and non-stop handlers in tree order.
	for i := range dt.entries {
		if dt.entries[i].matches(ke, fm) {
			debug.Log("dispatchTable.dispatch: normal handler fired at position %d, pattern=%+v, stop=%v", dt.entries[i].position, dt.entries[i].pattern, dt.entries[i].stop)
			dt.entries[i].handler(ke)
			if dt.entries[i].stop {
				debug.Log("dispatchTable.dispatch: stop handler consumed event")
				return true
			}
		}
	}
	debug.Log("dispatchTable.dispatch: no stop handler matched, returning false")
	return false
}

// validate checks for conflicting Stop handlers. Two active Stop handlers
// for the same key pattern is an error — it's ambiguous which should win.
// A Stop handler + a broadcast handler for the same pattern is fine.
func (dt *dispatchTable) validate() error {
	// Track patterns that already have a Stop handler
	type stopInfo struct {
		position int
	}
	stopPatterns := make(map[KeyPattern]stopInfo)

	for _, entry := range dt.entries {
		if !entry.stop {
			continue
		}
		// Focus-gated entries cannot conflict because only one can be focused at a time
		if entry.pattern.FocusRequired {
			continue
		}
		// Strip FocusRequired for comparison so focus-gated and broadcast entries
		// with the same key don't conflict
		comparePattern := entry.pattern
		comparePattern.FocusRequired = false
		if existing, conflict := stopPatterns[comparePattern]; conflict {
			return fmt.Errorf(
				"conflicting stop handlers for key pattern %+v at tree positions %d and %d",
				entry.pattern, existing.position, entry.position,
			)
		}
		stopPatterns[comparePattern] = stopInfo{position: entry.position}
	}

	return nil
}
