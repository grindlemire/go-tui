package tui

import "fmt"

// FocusGroup manages Tab/Shift+Tab cycling between a set of components.
// Each member is a *State[bool] that indicates whether that component is
// "selected" (active). FocusGroup ensures mutual exclusion: exactly one
// member is active at a time.
//
// FocusGroup implements KeyListener but not Component — it is a helper
// that participates in the key dispatch system without rendering anything.
//
// Usage:
//
//	active1 := tui.NewState(true)  // First member starts active
//	active2 := tui.NewState(false)
//	fg := tui.NewFocusGroup(active1, active2)
//	// fg.KeyMap() returns Tab → Next, Shift+Tab → Prev
type FocusGroup struct {
	members []*State[bool]
	current int
}

// Compile-time interface check.
var _ KeyListener = (*FocusGroup)(nil)

// NewFocusGroup creates a FocusGroup managing the given members.
// The first member is initially active (set to true); all others are set to false.
// Returns an error if called with fewer than 2 members.
func NewFocusGroup(members ...*State[bool]) (*FocusGroup, error) {
	if len(members) < 2 {
		return nil, fmt.Errorf("focus group requires at least 2 members")
	}

	// Initialize: first member active, rest inactive.
	// Use setDirect because no app is bound yet during construction.
	for i, m := range members {
		m.setDirect(i == 0)
	}

	return &FocusGroup{
		members: members,
		current: 0,
	}, nil
}

// MustNewFocusGroup creates a FocusGroup and panics on error.
func MustNewFocusGroup(members ...*State[bool]) *FocusGroup {
	fg, err := NewFocusGroup(members...)
	if err != nil {
		panic(err)
	}
	return fg
}

// Next deactivates the current member and activates the next one (wrapping).
func (fg *FocusGroup) Next() {
	if len(fg.members) == 0 {
		return
	}
	fg.members[fg.current].Set(false)
	fg.current = (fg.current + 1) % len(fg.members)
	fg.members[fg.current].Set(true)
}

// Prev deactivates the current member and activates the previous one (wrapping).
func (fg *FocusGroup) Prev() {
	if len(fg.members) == 0 {
		return
	}
	fg.members[fg.current].Set(false)
	fg.current = fg.current - 1
	if fg.current < 0 {
		fg.current = len(fg.members) - 1
	}
	fg.members[fg.current].Set(true)
}

// Current returns the index of the currently active member.
func (fg *FocusGroup) Current() int {
	return fg.current
}

// KeyMap returns key bindings for Tab (next) and Shift+Tab (prev).
// Tab is matched as KeyTab; Shift+Tab is matched as KeyTab with ModShift.
func (fg *FocusGroup) KeyMap() KeyMap {
	return KeyMap{
		{
			Pattern: KeyPattern{Key: KeyTab, RequireNoMods: true},
			Handler: func(ke KeyEvent) { fg.Next() },
			Stop:    false,
		},
		{
			Pattern: KeyPattern{Key: KeyTab, Mod: ModShift},
			Handler: func(ke KeyEvent) { fg.Prev() },
			Stop:    false,
		},
	}
}
