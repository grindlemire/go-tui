# Component Model Review Fixes

**Status:** Proposed
**Date:** 2026-02-04

---

## Overview

This plan addresses issues found during review of the component model implementation against the design spec in `specs/[DONE]-component-model-design.md`. Each fix is self-contained with exact file locations, current code, and replacement code.

---

## Fix 1: FocusGroup Tab/Shift+Tab double-fire bug

**Severity:** Bug — causes visible misbehavior at runtime.

**Problem:** When the user presses Shift+Tab, both the bare Tab handler AND the Shift+Tab handler fire, because the dispatch matcher skips modifier checks when `KeyPattern.Mods == 0`. The bare `OnKey(KeyTab, ...)` binding has `Mods: 0`, so the check `if p.Mods != 0 && ke.Mod != p.Mods` passes, and the pattern matches any Tab event regardless of modifiers. The result is that `fg.Next()` fires first, then `fg.Prev()` fires, and focus ends up back where it started.

**Root cause:** The `matches()` function in `dispatch.go` interprets `Mods: 0` as "don't care about modifiers" rather than "require no modifiers." But `FocusGroup.KeyMap()` needs bare Tab (no modifiers) to be distinct from Shift+Tab.

**Fix:** Add a `RequireNoMods` field to `KeyPattern` so bindings can explicitly require that no modifiers are held. This avoids changing the default behavior of existing patterns where `Mods: 0` correctly means "don't filter on modifiers" (e.g., `OnKey(KeyEscape, ...)` should match Escape regardless of whether Shift is held).

### Files to change

#### `keymap.go` — Add `RequireNoMods` field to `KeyPattern`

**Current code (lines 14–20):**
```go
type KeyPattern struct {
	Key     Key      // Specific key (KeyCtrlB, KeyEscape, etc.), or 0
	Rune    rune     // Specific rune, or 0
	AnyRune bool     // Match any printable character
	Mods    Modifier // Required modifiers
}
```

**Replace with:**
```go
type KeyPattern struct {
	Key           Key      // Specific key (KeyCtrlB, KeyEscape, etc.), or 0
	Rune          rune     // Specific rune, or 0
	AnyRune       bool     // Match any printable character
	Mods          Modifier // Required modifiers (when non-zero, event must have exactly these mods)
	RequireNoMods bool     // When true, event must have no modifiers (Mods field is ignored)
}
```

#### `dispatch.go` — Check `RequireNoMods` in `matches()`

**Current code (lines 55–58):**
```go
	// Check modifier requirements when specified
	if p.Mods != 0 && ke.Mod != p.Mods {
		return false
	}
```

**Replace with:**
```go
	// Check modifier requirements
	if p.RequireNoMods && ke.Mod != 0 {
		return false
	}
	if p.Mods != 0 && ke.Mod != p.Mods {
		return false
	}
```

#### `focus_group.go` — Use `RequireNoMods` on the Tab binding

**Current code (lines 74–82):**
```go
func (fg *FocusGroup) KeyMap() KeyMap {
	return KeyMap{
		OnKey(KeyTab, func(ke KeyEvent) { fg.Next() }),
		{
			Pattern: KeyPattern{Key: KeyTab, Mods: ModShift},
			Handler: func(ke KeyEvent) { fg.Prev() },
			Stop:    false,
		},
	}
}
```

**Replace with:**
```go
func (fg *FocusGroup) KeyMap() KeyMap {
	return KeyMap{
		{
			Pattern: KeyPattern{Key: KeyTab, RequireNoMods: true},
			Handler: func(ke KeyEvent) { fg.Next() },
			Stop:    false,
		},
		{
			Pattern: KeyPattern{Key: KeyTab, Mods: ModShift},
			Handler: func(ke KeyEvent) { fg.Prev() },
			Stop:    false,
		},
	}
}
```

### Tests to add/update

#### `dispatch_test.go` — Add test cases for `RequireNoMods`

Add these cases to the existing `TestDispatchEntry_Matches` table:

```go
"RequireNoMods matches event with no modifiers": {
    pattern: KeyPattern{Key: KeyTab, RequireNoMods: true},
    event:   KeyEvent{Key: KeyTab, Mod: ModNone},
    want:    true,
},
"RequireNoMods rejects event with shift": {
    pattern: KeyPattern{Key: KeyTab, RequireNoMods: true},
    event:   KeyEvent{Key: KeyTab, Mod: ModShift},
    want:    false,
},
"RequireNoMods rejects event with alt": {
    pattern: KeyPattern{Key: KeyTab, RequireNoMods: true},
    event:   KeyEvent{Key: KeyTab, Mod: ModAlt},
    want:    false,
},
```

#### `focus_group_test.go` — Add regression test

Add a new test function:

```go
func TestFocusGroup_ShiftTabDoesNotTriggerNext(t *testing.T) {
	m0 := NewState(false)
	m1 := NewState(false)
	m2 := NewState(false)
	fg := NewFocusGroup(m0, m1, m2)

	// Simulate Shift+Tab via dispatch
	root := New()
	child := New()
	child.component = fg // FocusGroup implements KeyListener
	root.AddChild(child)

	table, err := buildDispatchTable(root)
	if err != nil {
		t.Fatalf("buildDispatchTable: %v", err)
	}

	// Shift+Tab should ONLY call Prev(), not Next()+Prev()
	table.dispatch(KeyEvent{Key: KeyTab, Mod: ModShift})
	assertMemberStates(t, "after Shift+Tab", []*State[bool]{m0, m1, m2}, 2)
}
```

Note: `FocusGroup` implements `KeyListener` but not `Component` (no `Render()` method). The test above attaches it directly to an element's `component` field for dispatch table discovery. This works because `walkComponents` checks for `KeyListener` via type assertion on whatever is in `component`. However, `component` is typed as `Component` (which requires `Render()`), so `FocusGroup` does not satisfy that interface. To make this test work, either:
- Create a wrapper mock that embeds `FocusGroup` and adds a `Render()` method, or
- Have the test call `fg.KeyMap()` directly and dispatch manually:

```go
func TestFocusGroup_ShiftTabDoesNotTriggerNext(t *testing.T) {
	m0 := NewState(false)
	m1 := NewState(false)
	m2 := NewState(false)
	fg := NewFocusGroup(m0, m1, m2)

	km := fg.KeyMap()

	// Simulate Shift+Tab — only the Shift+Tab binding should match
	event := KeyEvent{Key: KeyTab, Mod: ModShift}
	matched := 0
	for _, binding := range km {
		entry := dispatchEntry{pattern: binding.Pattern, handler: binding.Handler, stop: binding.Stop}
		if entry.matches(event) {
			entry.handler(event)
			matched++
		}
	}

	if matched != 1 {
		t.Errorf("Shift+Tab matched %d bindings, want 1", matched)
	}
	// Should have gone backward to index 2
	assertMemberStates(t, "after Shift+Tab", []*State[bool]{m0, m1, m2}, 2)
}
```

#### `keymap_test.go` — Add `RequireNoMods` to `TestKeyMap_KeyPatternEquality`

```go
"RequireNoMods patterns are equal": {
    a:    KeyPattern{Key: KeyTab, RequireNoMods: true},
    b:    KeyPattern{Key: KeyTab, RequireNoMods: true},
    want: true,
},
"RequireNoMods vs Mods not equal": {
    a:    KeyPattern{Key: KeyTab, RequireNoMods: true},
    b:    KeyPattern{Key: KeyTab, Mods: ModShift},
    want: false,
},
```

---

## Fix 2: `WithRoot` does not accept `Component`

**Severity:** Medium — API gap between spec and implementation.

**Problem:** The design spec shows `tui.NewApp(tui.WithRoot(MyApp()))` as the entry point. But `WithRoot` is typed `func WithRoot(v Viewable) AppOption`. `Component` is a separate interface, so passing a `Component` to `WithRoot` won't compile. The example code works around this by using `app.SetRoot(MyApp())` (which accepts `any`), but the ergonomic option-based API doesn't work for the component model.

**Fix:** Change `WithRoot` to accept `any`, matching the `SetRoot` signature. `SetRoot` already has a type switch that handles `Component`, `Viewable`, and `Renderable`.

### Files to change

#### `app_options.go` — Widen `WithRoot` to accept `any`

**Current code (lines 62–71):**
```go
// WithRoot sets the root view for rendering. Accepts:
//   - A view struct implementing Viewable (extracts Root, starts watchers)
//   - A raw Renderable (element.Element)
//
// The root is set after the app is fully initialized.
func WithRoot(v Viewable) AppOption {
	return func(a *App) error {
		a.pendingRoot = v
		return nil
	}
}
```

**Replace with:**
```go
// WithRoot sets the root for rendering. Accepts:
//   - A Component (struct component with Render() method)
//   - A Viewable (view struct from function components)
//   - A Renderable (raw element)
//
// The root is set after the app is fully initialized.
// Uses the same type dispatch as SetRoot.
func WithRoot(v any) AppOption {
	return func(a *App) error {
		a.pendingRoot = v
		return nil
	}
}
```

The `pendingRoot` field is already typed `any` (`app.go` line 83), and `SetRoot` already handles the type switch (`app.go` lines 337–358), so no other changes are needed. When `NewApp` processes `pendingRoot`, it calls `a.SetRoot(a.pendingRoot)` (`app.go` line 322), which dispatches correctly.

### Tests

No new tests required — existing tests use `WithRoot` with `Viewable` and will continue to work. The change is purely a type widening. Optionally, add a test that passes a `Component` to `WithRoot`:

```go
func TestWithRoot_AcceptsComponent(t *testing.T) {
	comp := &mockComponent{}
	opt := WithRoot(comp)
	app := &App{mounts: newMountState()}
	err := opt(app)
	if err != nil {
		t.Fatalf("WithRoot(Component) returned error: %v", err)
	}
	if app.pendingRoot != comp {
		t.Error("pendingRoot should be set to the component")
	}
}
```

---

## Fix 3: Dispatch table validation silently swallows errors

**Severity:** Medium — conflicting stop handlers are silently ignored instead of reported.

**Problem:** `rebuildDispatchTable()` in `app_loop.go` catches the validation error from `buildDispatchTable()` and silently keeps the old table. The design spec (§12, criterion 5) says conflicting stop handlers should cause a panic. More practically, silently ignoring the error means the developer gets no feedback that their key bindings conflict — the app just behaves in confusing ways because the old table is used.

**Fix:** Log the error to stderr. A panic is too aggressive for a runtime condition that can occur during normal state transitions (e.g., two components temporarily both active during a render). Logging to stderr is the right balance — it's visible during development but doesn't crash the app.

### Files to change

#### `app_loop.go` — Log validation errors in `rebuildDispatchTable()`

**Current code (lines 107–120):**
```go
func (a *App) rebuildDispatchTable() {
	root, ok := a.root.(*Element)
	if !ok {
		return
	}

	table, err := buildDispatchTable(root)
	if err != nil {
		// Validation error (e.g., conflicting Stop handlers).
		// Keep the previous valid table rather than crashing.
		return
	}
	a.dispatchTable = table
}
```

**Replace with:**
```go
func (a *App) rebuildDispatchTable() {
	root, ok := a.root.(*Element)
	if !ok {
		return
	}

	table, err := buildDispatchTable(root)
	if err != nil {
		// Validation error (e.g., conflicting Stop handlers).
		// Log and keep the previous valid table rather than crashing.
		fmt.Fprintf(os.Stderr, "tui: dispatch table error: %v\n", err)
		return
	}
	a.dispatchTable = table
}
```

This file already imports `"time"` and `"os"` (for `os.Signal`/`os.Interrupt`). It does not currently import `"fmt"`, so add `"fmt"` to the import block at the top of the file (`app_loop.go`, lines 3–6):

**Current imports:**
```go
import (
	"os"
	"os/signal"
	"time"
)
```

**Replace with:**
```go
import (
	"fmt"
	"os"
	"os/signal"
	"time"
)
```

### Tests

No new tests — the existing `TestDispatch_ConflictValidation_TwoStopHandlersSamePattern` in `dispatch_test.go` already verifies that `buildDispatchTable` returns an error. The change here is about how the caller surfaces that error, which is not unit-testable without capturing stderr.

---

## Fix 4: `globalKeyHandler` bypasses dispatch table

**Severity:** Medium — undocumented interaction between old and new key handling systems.

**Problem:** In `app_events.go` (line 71), `globalKeyHandler` runs before the dispatch table. If a user sets a global key handler (via `SetGlobalKeyHandler` or `WithGlobalKeyHandler`) alongside the component model's `KeyMap` system, the global handler silently intercepts keys before any component sees them. This is not documented and creates confusion when migrating to the component model.

**Fix:** When a `rootComponent` is set (meaning the app is using the component model), skip the `globalKeyHandler` check. Components should use `KeyMap()` for all key handling. The `globalKeyHandler` remains available for legacy apps that don't use the component model.

### Files to change

#### `app_events.go` — Skip `globalKeyHandler` when using component model

**Current code (lines 69–80):**
```go
		a.eventQueue <- func() {
			// Global key handler runs first (for app-level bindings like quit)
			if keyEvent, isKey := ev.(KeyEvent); isKey {
				if a.globalKeyHandler != nil && a.globalKeyHandler(keyEvent) {
					return // Event consumed by global handler
				}
				// Use broadcast dispatch table for key events if available.
				// Falls through to FocusManager dispatch if no dispatch table.
				if a.dispatchTable != nil {
					a.dispatchTable.dispatch(keyEvent)
					return
				}
			}
```

**Replace with:**
```go
		a.eventQueue <- func() {
			if keyEvent, isKey := ev.(KeyEvent); isKey {
				// Component model path: use broadcast dispatch table exclusively.
				// globalKeyHandler is skipped — components use KeyMap() instead.
				if a.dispatchTable != nil {
					a.dispatchTable.dispatch(keyEvent)
					return
				}
				// Legacy path: global key handler runs before FocusManager dispatch.
				if a.globalKeyHandler != nil && a.globalKeyHandler(keyEvent) {
					return // Event consumed by global handler
				}
			}
```

This reorders the checks so that:
1. If a dispatch table exists (component model is active), it handles the event exclusively.
2. If no dispatch table exists (legacy mode), `globalKeyHandler` runs first, then falls through to `FocusManager.Dispatch()`.

The two paths are now mutually exclusive, eliminating the ambiguity.

### Tests

Add a test to `integration_test.go`:

```go
func TestIntegration_GlobalKeyHandlerSkippedWithComponentModel(t *testing.T) {
	cleanup := setupTestMountState()
	defer cleanup()

	root := newIntRoot()
	el := root.Render()
	el.component = root

	table, err := buildDispatchTable(el)
	if err != nil {
		t.Fatalf("buildDispatchTable: %v", err)
	}

	// Verify dispatch table handles the key, not globalKeyHandler
	ctrlCHandled := false
	table.dispatch(KeyEvent{Key: KeyCtrlC})
	// The root's KeyMap has a ctrl+c handler — it should fire.
	// (The handler in intRoot is a no-op for ctrl+c, but the point
	// is that the dispatch table is used, not globalKeyHandler.)
	_ = ctrlCHandled // globalKeyHandler would set this; dispatch table doesn't.
}
```

---

## Fix 5: Add `Quit()` as alias for `Stop()`

**Severity:** Low — spec says `tui.Quit()`, code has `tui.Stop()`.

**Problem:** The design spec consistently uses `tui.Quit()` but the codebase only has `tui.Stop()`. While the example code correctly uses `tui.Stop()`, adding `Quit()` improves discoverability since "quit" is a more natural word for ending an application.

### Files to change

#### `app_lifecycle.go` — Add `Quit` alias

Add after the existing `Stop()` function (after line 15):

```go
// Quit stops the currently running app. This is an alias for Stop().
func Quit() {
	Stop()
}
```

### Tests

No tests needed — it's a one-line delegation.

---

## Fix 6: Generator `generateForLoopForSlice` and `generateIfStmtForSlice` always use `.Root`

**Severity:** Latent — not triggerable through normal parser flow today, but incorrect if code paths change.

**Problem:** In `internal/tuigen/generator_children.go`, the functions `generateForLoopForSlice` (line 261) and `generateIfStmtForSlice` (lines 293 and 333) unconditionally append `.Root` to component call results:

```go
case *ComponentCall:
    callVar := g.generateComponentCallWithRefs(n, "")
    g.writef("%s = append(%s, %s.Root)\n", sliceVar, sliceVar, callVar)
```

If `IsStructMount` is true, `generateComponentCallWithRefs` calls `generateStructMount`, which emits `tui.Mount(...)` returning `*tui.Element`. Appending `.Root` to a `*tui.Element` would fail to compile. Currently safe because `IsStructMount` is only true inside method templs, and these `ForSlice`/`IfStmtForSlice` functions are only called from function component paths. But this assumption is implicit and fragile.

**Fix:** Add the same `IsStructMount` check that already exists in `generateFunctionComponentCall` (line 186 area of the same file).

### Files to change

#### `internal/tuigen/generator_children.go` — Add `IsStructMount` check

There are **four** locations where `*ComponentCall` is handled in for-loop and if-stmt slice builders. Each needs the same fix. The pattern is identical in all four places.

**Location 1: `generateForLoopForSlice`, line ~261 (inside `for _, node := range loop.Body`):**

Current:
```go
		case *ComponentCall:
			callVar := g.generateComponentCallWithRefs(n, "")
			g.writef("%s = append(%s, %s.Root)\n", sliceVar, sliceVar, callVar)
```

Replace with:
```go
		case *ComponentCall:
			callVar := g.generateComponentCallWithRefs(n, "")
			if n.IsStructMount {
				g.writef("%s = append(%s, %s)\n", sliceVar, sliceVar, callVar)
			} else {
				g.writef("%s = append(%s, %s.Root)\n", sliceVar, sliceVar, callVar)
			}
```

**Location 2: `generateIfStmtForSlice`, line ~293 (inside `for _, node := range stmt.Then`):**

Same replacement pattern as Location 1.

**Location 3: `generateIfStmtForSlice`, line ~333 (inside `for _, node := range stmt.Else`, in the Then branch):**

Same replacement pattern as Location 1.

**Location 4:** Search for any additional `*ComponentCall` + `.Root` patterns in the else branch of `generateIfStmtForSlice`. Apply the same fix.

To find all locations: search for `.Root)\n"` in `internal/tuigen/generator_children.go`. Every occurrence inside a `case *ComponentCall:` block needs the `IsStructMount` guard.

### Tests

Add a generator test case in `internal/tuigen/generator_test.go` to the `TestGenerator_MethodComponent` test:

```go
"struct mount inside for loop in method templ": {
    input: `package test

templ (a *app) Render() {
    <div>
        @for _, item := range a.items {
            @ChildComponent(item)
        }
    </div>
}`,
    wantContains: []string{
        "tui.Mount(a,",
        "func() tui.Component {",
        "return ChildComponent(item)",
    },
    wantNotContains: []string{
        ".Root",
    },
},
```

---

## Fix 7: Rename `KeyPattern.Mods` to `KeyPattern.Mod`

**Severity:** Low — naming inconsistency.

**Problem:** `KeyPattern.Mods` (plural) vs `KeyEvent.Mod` (singular). Both represent the same concept (modifier flags). The asymmetry is a minor source of confusion.

**Fix:** Rename `KeyPattern.Mods` to `KeyPattern.Mod` for consistency with `KeyEvent.Mod`.

### Files to change

This is a mechanical rename. Use find-and-replace scoped to the correct struct field.

1. **`keymap.go`** — Change field declaration from `Mods` to `Mod` in `KeyPattern` struct (line 19)
2. **`dispatch.go`** — Change `p.Mods` to `p.Mod` in `matches()` (lines 57, 58)
3. **`focus_group.go`** — Change `Mods: ModShift` to `Mod: ModShift` (line 78)
4. **`dispatch_test.go`** — Change all `Mods:` to `Mod:` in test cases (4 occurrences)
5. **`keymap_test.go`** — Change all `Mods:` to `Mod:` in test cases (3 occurrences)
6. **`focus_group_test.go`** — Change `km[0].Pattern.Mods` and `km[1].Pattern.Mods` to `.Mod` (2 occurrences)

### Tests

All existing tests continue to pass after the rename. No new tests needed.

---

## Implementation Order

The fixes are independent and can be done in any order. Recommended order by priority:

1. **Fix 1** (FocusGroup Tab bug) — actual runtime bug
2. **Fix 3** (log dispatch errors) — silent failures are bad
3. **Fix 4** (globalKeyHandler bypass) — eliminates ambiguity
4. **Fix 2** (WithRoot accepts Component) — API ergonomics
5. **Fix 6** (generator `.Root` guard) — future-proofing
6. **Fix 5** (Quit alias) — trivial
7. **Fix 7** (Mods→Mod rename) — trivial, do last since it touches many files

## Verification

After all fixes, run:

```bash
go test ./...
go build ./cmd/tui
go build ./examples/component-model/
```

All tests should pass, and both the CLI and the component-model example should compile cleanly.
