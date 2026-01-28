package tui

import "sync/atomic"

// dirty is the global dirty flag - set by any mutating operation.
// When true, the UI needs to be re-rendered.
var dirty atomic.Bool

// MarkDirty signals that the UI needs re-rendering.
// Called automatically by State.Set(), Element.ScrollBy(), etc.
// Can also be called manually for custom mutations.
func MarkDirty() {
	dirty.Store(true)
}

// checkAndClearDirty returns true if dirty and clears the flag.
// Called by the main loop after processing events.
func checkAndClearDirty() bool {
	return dirty.Swap(false)
}

// resetDirty clears the dirty flag without returning its value.
// Used for testing to reset state between tests.
func resetDirty() {
	dirty.Store(false)
}

// TestCheckAndClearDirty is an exported version of checkAndClearDirty for testing.
// Only use this in test code.
func TestCheckAndClearDirty() bool {
	return checkAndClearDirty()
}

// TestResetDirty is an exported version of resetDirty for testing.
// Only use this in test code.
func TestResetDirty() {
	resetDirty()
}
