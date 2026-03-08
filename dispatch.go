package tui

import "fmt"

// dispatchEntry is a handler with its tree position for ordering.
type dispatchEntry struct {
	pattern   KeyPattern
	handler   func(KeyEvent)
	stop      bool
	position  int       // BFS order index from tree walk
	focusable Focusable // Non-nil for focus-gated entries; checked at dispatch time
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

		// Check if this component implements Focusable (for focus-gated bindings)
		focusableComp, _ := comp.(Focusable)

		for _, binding := range km {
			entry := dispatchEntry{
				pattern:  binding.Pattern,
				handler:  binding.Handler,
				stop:     binding.Stop,
				position: position,
			}
			// Only store focusable ref if the binding requires focus
			if binding.Pattern.FocusRequired && focusableComp != nil {
				entry.focusable = focusableComp
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

// matches checks if a dispatch entry matches a key event.
func (e *dispatchEntry) matches(ke KeyEvent, fm *focusManager) bool {
	p := e.pattern

	// Focus-gated: skip if not focused
	if p.FocusRequired && e.focusable != nil {
		if fm == nil || !fm.IsFocused(e.focusable) {
			return false
		}
	}

	// Check modifier requirements
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

// dispatch sends a key event to all matching handlers in tree order.
// Stops early if a matching handler has Stop=true.
// Returns true if a handler with Stop=true consumed the event.
func (dt *dispatchTable) dispatch(ke KeyEvent, fm *focusManager) bool {
	if dt == nil {
		return false
	}
	for i := range dt.entries {
		if dt.entries[i].matches(ke, fm) {
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
		position int
	}
	stopPatterns := make(map[KeyPattern]stopInfo)

	for _, entry := range dt.entries {
		if !entry.stop {
			continue
		}
		if existing, conflict := stopPatterns[entry.pattern]; conflict {
			return fmt.Errorf(
				"conflicting stop handlers for key pattern %+v at tree positions %d and %d",
				entry.pattern, existing.position, entry.position,
			)
		}
		stopPatterns[entry.pattern] = stopInfo{position: entry.position}
	}

	return nil
}
