# Focus Reference

## Overview

Focus determines which element receives keyboard input. go-tui provides three layers of focus management:

- **Element-level** -- individual elements marked as focusable, with `Focus()` and `Blur()` methods
- **focusManager** -- internal tracking used by `App` to cycle through focusable elements with `FocusNext()` and `FocusPrev()`
- **FocusGroup** -- a state-driven helper for Tab/Shift+Tab cycling between logical sections of your UI

Most applications use element-level focus through `.gsx` attributes and `App`-level navigation. `FocusGroup` is useful when you need to toggle between distinct panels or sections rather than individual elements.

## Focusable Interface

```go
type Focusable interface {
    IsFocusable() bool
    HandleEvent(event Event) bool
    Focus()
    Blur()
}
```

`Focusable` is the contract for anything that can receive keyboard focus. `Element` satisfies this interface directly, so you rarely need to implement it yourself.

### IsFocusable

```go
IsFocusable() bool
```

Returns whether this element can currently receive focus. Returns the value set by `WithFocusable` or `SetFocusable` (defaults to `false`).

### HandleEvent

```go
HandleEvent(event Event) bool
```

Processes a keyboard or mouse event. Returns `true` if the event was consumed (preventing further propagation), `false` otherwise. For scrollable elements, `Element.HandleEvent` automatically handles arrow keys, Page Up/Down, Home/End, and mouse wheel events.

### Focus

```go
Focus()
```

Called when the element gains focus. Implementations typically update visual state -- for example, changing a border color or adding a highlight. `Element.Focus` sets an internal `focused` flag and fires the `onFocus` callback if one was registered.

### Blur

```go
Blur()
```

Called when the element loses focus. Reverses whatever visual change `Focus` applied. `Element.Blur` clears the internal `focused` flag and fires the `onBlur` callback if one was registered.

## Element Focus API

`Element` implements `Focusable` and adds several focus-related methods and options.

### Option Functions

#### WithFocusable

```go
func WithFocusable(focusable bool) Option
```

Sets whether the element can receive focus. Elements are not focusable by default.

In `.gsx`:

```gsx
<div focusable={true}>
    <span>I can receive focus</span>
</div>
```

#### WithOnFocus

```go
func WithOnFocus(fn func(*Element)) Option
```

Sets a callback that fires when the element gains focus. The callback receives the element itself. Implicitly sets `focusable = true`.

In `.gsx`:

```gsx
<div onFocus={s.handleFocus}>
    <span>Focus me</span>
</div>
```

#### WithOnBlur

```go
func WithOnBlur(fn func(*Element)) Option
```

Sets a callback that fires when the element loses focus. The callback receives the element itself. Implicitly sets `focusable = true`.

In `.gsx`:

```gsx
<div onBlur={s.handleBlur}>
    <span>Blur me</span>
</div>
```

### Methods

#### IsFocused

```go
func (e *Element) IsFocused() bool
```

Returns whether this element currently has focus. Use in render methods to apply conditional styling.

```gsx
// Helper function for focus-dependent styling
func panelBorderStyle(ref *tui.Ref) tui.Style {
    if ref.El() != nil && ref.El().IsFocused() {
        return tui.NewStyle().Foreground(tui.ANSIColor(tui.Cyan))
    }
    return tui.NewStyle().Foreground(tui.ANSIColor(tui.White))
}

templ (s *myComp) Render() {
    <div ref={s.panel} focusable={true} borderStyle={panelBorderStyle(s.panel)} class="border-rounded p-1">
        <span>Panel content</span>
    </div>
}
```

#### SetFocusable

```go
func (e *Element) SetFocusable(focusable bool)
```

Changes whether the element can receive focus at runtime.

#### SetOnFocus

```go
func (e *Element) SetOnFocus(fn func(*Element))
```

Sets the focus callback at runtime. Implicitly sets `focusable = true`.

#### SetOnBlur

```go
func (e *Element) SetOnBlur(fn func(*Element))
```

Sets the blur callback at runtime. Implicitly sets `focusable = true`.

#### ContainsPoint

```go
func (e *Element) ContainsPoint(x, y int) bool
```

Returns `true` if the point `(x, y)` falls within the element's layout bounds. Useful for mouse-based hit testing in `HandleMouse` implementations.

### Focus Tree Discovery

The `App` uses these methods internally to find and register focusable elements. You normally don't need to call them directly.

#### WalkFocusables

```go
func (e *Element) WalkFocusables(fn func(Focusable))
```

Walks the element tree depth-first, calling `fn` for each focusable element. Skips hidden elements and their subtrees. The `App` calls this after each render to discover new focusable elements.

#### SetOnFocusableAdded

```go
func (e *Element) SetOnFocusableAdded(fn func(Focusable))
```

Sets a callback that fires when a focusable descendant is added to the tree. The `App` uses this to auto-register focusable elements as they appear.

## App Focus Methods

The `App` delegates focus navigation to its internal focus manager through three methods.

### FocusNext

```go
func (a *App) FocusNext()
```

Moves focus to the next focusable element in document order (depth-first traversal of the element tree). Wraps around to the first element when reaching the end.

```go
func (s *myApp) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.OnKey(tui.KeyTab, func(ke tui.KeyEvent) {
            ke.App().FocusNext()
        }),
    }
}
```

### FocusPrev

```go
func (a *App) FocusPrev()
```

Moves focus to the previous focusable element. Wraps around to the last element when reaching the beginning.

```go
tui.OnKey(tui.KeyTab, func(ke tui.KeyEvent) {
    if ke.Mod.Has(tui.ModShift) {
        ke.App().FocusPrev()
    } else {
        ke.App().FocusNext()
    }
})
```

### Focused

```go
func (a *App) Focused() Focusable
```

Returns the currently focused element, or `nil` if nothing has focus.

```go
if focused := app.Focused(); focused != nil {
    // Something has focus
}
```

## FocusGroup

`FocusGroup` manages Tab/Shift+Tab cycling between logical sections of a UI. Each section is represented by a `*State[bool]` that indicates whether that section is active. `FocusGroup` enforces mutual exclusion: exactly one member is active at a time.

`FocusGroup` implements `KeyListener` (it has a `KeyMap()` method) but is not a `Component`. It participates in the key dispatch system without rendering anything.

### NewFocusGroup

```go
func NewFocusGroup(members ...*State[bool]) (*FocusGroup, error)
```

Creates a `FocusGroup` managing the given members. The constructor always activates the first member (`true`) and deactivates the rest (`false`), regardless of their initial values. Returns an error if fewer than 2 members are provided.

```go
panelA := tui.NewState(false)
panelB := tui.NewState(false)
panelC := tui.NewState(false)

fg, err := tui.NewFocusGroup(panelA, panelB, panelC)
// panelA is now true, panelB and panelC are false
if err != nil {
    // handle error
}
```

### MustNewFocusGroup

```go
func MustNewFocusGroup(members ...*State[bool]) *FocusGroup
```

Same as `NewFocusGroup` but panics on error. Use when the member count is known at compile time.

```go
fg := tui.MustNewFocusGroup(panelA, panelB, panelC)
```

### Next

```go
func (fg *FocusGroup) Next()
```

Deactivates the current member (sets its state to `false`) and activates the next one (sets to `true`), wrapping from the last member back to the first.

### Prev

```go
func (fg *FocusGroup) Prev()
```

Deactivates the current member and activates the previous one, wrapping from the first member to the last.

### Current

```go
func (fg *FocusGroup) Current() int
```

Returns the zero-based index of the currently active member.

### KeyMap

```go
func (fg *FocusGroup) KeyMap() KeyMap
```

Returns two key bindings:

- **Tab** (no modifiers) calls `Next()`
- **Shift+Tab** calls `Prev()`

Both bindings use `Stop: false`, so the events continue propagating after the focus group handles them.

```go
func (s *myApp) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        // FocusGroup bindings first
        s.focusGroup.KeyMap()[0],
        s.focusGroup.KeyMap()[1],
        // Then app-level bindings
        tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) {
            ke.App().Stop()
        }),
    }
}
```

Or spread the entire KeyMap:

```go
func (s *myApp) KeyMap() tui.KeyMap {
    km := s.focusGroup.KeyMap()
    km = append(km,
        tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) {
            ke.App().Stop()
        }),
    )
    return km
}
```

## Complete Example

A three-panel layout with FocusGroup-driven Tab cycling. Each panel highlights its border when active.

```gsx
package main

import (
    "fmt"

    tui "github.com/grindlemire/go-tui"
)

type panels struct {
    panelA     *tui.State[bool]
    panelB     *tui.State[bool]
    panelC     *tui.State[bool]
    focusGroup *tui.FocusGroup
}

func Panels() *panels {
    a := tui.NewState(false)
    b := tui.NewState(false)
    c := tui.NewState(false)
    return &panels{
        panelA:     a,
        panelB:     b,
        panelC:     c,
        focusGroup: tui.MustNewFocusGroup(a, b, c),
    }
}

func (p *panels) KeyMap() tui.KeyMap {
    km := p.focusGroup.KeyMap()
    km = append(km,
        tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) {
            ke.App().Stop()
        }),
    )
    return km
}

func panelStyle(active bool) string {
    if active {
        return "text-cyan font-bold"
    }
    return "font-dim"
}

func panelBorder(active bool) tui.Style {
    if active {
        return tui.NewStyle().Foreground(tui.ANSIColor(tui.Cyan))
    }
    return tui.NewStyle().Foreground(tui.ANSIColor(tui.White))
}

templ (p *panels) Render() {
    <div class="flex-row gap-1 p-1" height={20}>
        <Panel title="Panel A" active={p.panelA.Get()} />
        <Panel title="Panel B" active={p.panelB.Get()} />
        <Panel title="Panel C" active={p.panelC.Get()} />
    </div>
}

templ Panel(title string, active bool) {
    <div class="flex-col flex-1 border-rounded p-1" borderStyle={panelBorder(active)}>
        <span class={panelStyle(active)}>{title}</span>
        @if active {
            <span class="text-cyan">{fmt.Sprintf("%s is active", title)}</span>
        } @else {
            <span class="text-dim">Press Tab to focus</span>
        }
    </div>
}
```

## See Also

- [Events Reference](events.md) -- key and mouse event types
- [State Reference](state.md) -- `State[T]` used by FocusGroup members
- [App Reference](app.md) -- `FocusNext`, `FocusPrev`, `Focused` methods
- [Focus Guide](../guides/13-focus.md) -- practical patterns for focus management
