package tui

import "fmt"

// dispatchEntry is a handler with its tree position for ordering.
type dispatchEntry struct {
	pattern  KeyPattern
	handler  func(KeyEvent)
	stop     bool
	position int // DFS order index from tree walk
}

// dispatchTable holds all handlers in a single tree-ordered list.
// Handlers are matched against incoming KeyEvents by pattern.
type dispatchTable struct {
	entries []dispatchEntry // All handlers, ordered by tree position
}

// buildDispatchTable walks the element tree, collects KeyMap() from
// all mounted components, validates exclusive conflicts, and builds
// the dispatch table ordered by tree position.
func buildDispatchTable(root *Element) (*dispatchTable, error) {
	table := &dispatchTable{}
	position := 0

	walkComponents(root, func(comp Component) {
		kl, ok := comp.(KeyListener)
		if !ok {
			return
		}
		km := kl.KeyMap()
		if km == nil {
			return
		}
		for _, binding := range km {
			table.entries = append(table.entries, dispatchEntry{
				pattern:  binding.Pattern,
				handler:  binding.Handler,
				stop:     binding.Stop,
				position: position,
			})
		}
		position++
	})

	if err := table.validate(); err != nil {
		return nil, err
	}

	return table, nil
}

// matches checks if a dispatch entry matches a key event.
func (e *dispatchEntry) matches(ke KeyEvent) bool {
	p := e.pattern

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
func (dt *dispatchTable) dispatch(ke KeyEvent) {
	if dt == nil {
		return
	}
	for i := range dt.entries {
		if dt.entries[i].matches(ke) {
			dt.entries[i].handler(ke)
			if dt.entries[i].stop {
				return
			}
		}
	}
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
