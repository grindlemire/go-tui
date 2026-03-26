# Events Reference

## Overview

go-tui routes terminal input through a typed event system. Keyboard presses, mouse clicks, and terminal resizes each have their own event type. Components receive these events through two interfaces: `KeyListener` for keyboard input and `MouseListener` for mouse input.

The `KeyMap` system gives you a declarative way to bind keys to handlers. Bindings are checked in tree order, and you can control whether an event continues propagating or stops at the first match.

## Event Interface

```go
type Event interface {
    isEvent()
}
```

`Event` is a marker interface shared by all event types. The `isEvent()` method is unexported, so only the framework's own types satisfy it. Use a type switch to handle specific event types:

```go
switch ev := event.(type) {
case tui.KeyEvent:
    // handle keyboard input
case tui.MouseEvent:
    // handle mouse input
case tui.ResizeEvent:
    // handle terminal resize
}
```

## KeyEvent

```go
type KeyEvent struct {
    Key  Key
    Rune rune
    Mod  Modifier
}
```

Represents a single keyboard input.

**Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `Key` | `Key` | The key pressed. For printable characters this is `KeyRune`; for special keys (arrows, function keys, etc.) it's the specific constant. |
| `Rune` | `rune` | The character for `KeyRune` events. Zero for special keys. |
| `Mod` | `Modifier` | Modifier flags: any combination of `ModCtrl`, `ModAlt`, `ModShift`. |

**Methods:**

### App

```go
func (e KeyEvent) App() *App
```

Returns the `App` that dispatched this event. Use it in key handlers to call app-level methods.

```go
tui.On(tui.KeyEscape, func(ke tui.KeyEvent) {
    ke.App().Stop()
})
```

### IsRune

```go
func (e KeyEvent) IsRune() bool
```

Returns `true` if the event is a printable character (i.e., `Key == KeyRune`).

```go
tui.On(tui.AnyRune, func(ke tui.KeyEvent) {
    if ke.IsRune() {
        fmt.Printf("typed: %c\n", ke.Rune)
    }
})
```

### Is

```go
func (e KeyEvent) Is(key Key, mods ...Modifier) bool
```

Checks whether the event matches a key and (optionally) a set of modifiers. When multiple modifiers are passed, they are combined, and the event must have exactly that modifier set.

```go
// Match Enter with no modifier check
ke.Is(tui.KeyEnter)

// Match Ctrl+S
ke.Is(tui.KeyCtrlS)

// Match a key with specific modifiers
ke.Is(tui.KeyUp, tui.ModShift)
```

### Char

```go
func (e KeyEvent) Char() rune
```

Returns the rune if the event is a `KeyRune` event, or `0` otherwise. A convenience wrapper around checking `Key == KeyRune` and reading `Rune`.

## MouseEvent

```go
type MouseEvent struct {
    Button MouseButton
    Action MouseAction
    X      int
    Y      int
    Mod    Modifier
}
```

Represents a mouse input event. Mouse events are only delivered when mouse reporting is enabled (see [App Reference](app.md), `WithMouse()`).

**Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `Button` | `MouseButton` | Which button was involved. |
| `Action` | `MouseAction` | The type of action (press, release, drag). |
| `X` | `int` | Column position, 0-indexed from the left edge. |
| `Y` | `int` | Row position, 0-indexed from the top edge. |
| `Mod` | `Modifier` | Modifier flags held during the mouse event. |

**Methods:**

### App

```go
func (e MouseEvent) App() *App
```

Returns the `App` that dispatched this event.

## ResizeEvent

```go
type ResizeEvent struct {
    Width  int
    Height int
}
```

Emitted when the terminal window changes size. The framework handles resize internally: it updates the buffer dimensions, recalculates layout, and triggers a full redraw. You rarely need to handle this yourself.

## Key Constants

`Key` is a `uint16` enum. Every constant has a `String()` method that returns a human-readable name (e.g., `"Escape"`, `"Ctrl+A"`, `"F5"`).

### Special Keys

| Constant | String | Description |
|----------|--------|-------------|
| `KeyNone` | `"None"` | Zero value. No key. |
| `KeyRune` | `"Rune"` | Printable character. Check `Rune` field for the character. |
| `KeyEscape` | `"Escape"` | Escape key. |
| `KeyEnter` | `"Enter"` | Enter / Return key. |
| `KeyTab` | `"Tab"` | Tab key. |
| `KeyBackspace` | `"Backspace"` | Backspace key. |
| `KeyDelete` | `"Delete"` | Delete key. |
| `KeyInsert` | `"Insert"` | Insert key. |

### Arrow Keys

| Constant | String |
|----------|--------|
| `KeyUp` | `"Up"` |
| `KeyDown` | `"Down"` |
| `KeyLeft` | `"Left"` |
| `KeyRight` | `"Right"` |

### Navigation Keys

| Constant | String |
|----------|--------|
| `KeyHome` | `"Home"` |
| `KeyEnd` | `"End"` |
| `KeyPageUp` | `"PageUp"` |
| `KeyPageDown` | `"PageDown"` |

### Function Keys

| Constant | String |
|----------|--------|
| `KeyF1` – `KeyF12` | `"F1"` – `"F12"` |

All twelve function keys are defined: `KeyF1`, `KeyF2`, `KeyF3`, `KeyF4`, `KeyF5`, `KeyF6`, `KeyF7`, `KeyF8`, `KeyF9`, `KeyF10`, `KeyF11`, `KeyF12`.

### Control Keys

| Constant | String |
|----------|--------|
| `KeyCtrlA` – `KeyCtrlZ` | `"Ctrl+A"` – `"Ctrl+Z"` |
| `KeyCtrlSpace` | `"Ctrl+Space"` |

Each constant is a `RuneSpec` matching the corresponding `Rune(letter).Ctrl()` pattern. `On(tui.KeyCtrlS, handler)` and `On(tui.Rune('s').Ctrl(), handler)` are equivalent.

### Ctrl+H, Ctrl+I, Ctrl+M and Backspace/Tab/Enter

Three Ctrl+letter combinations share a terminal byte with a functional key:

| Ctrl combo | Functional key | Shared legacy byte |
|------------|---------------|-------------------|
| Ctrl+H | Backspace | `0x08` |
| Ctrl+I | Tab | `0x09` |
| Ctrl+M | Enter | `0x0D` |

**Ctrl+H / Backspace (0x08):** Modern terminals send `0x7F` for Backspace, so go-tui treats `0x08` as Ctrl+H. `KeyCtrlH` and `KeyBackspace` are separate bindings: Backspace matches `0x7F` (and Kitty's `CSI 127;1u`), while `KeyCtrlH` matches `0x08` (and Kitty's `CSI 104;5u`). Terminals configured with `stty erase ^H` send `0x08` for Backspace, which will fire `KeyCtrlH` handlers instead of `KeyBackspace`.

**Ctrl+I / Tab and Ctrl+M / Enter:** In legacy mode, the terminal sends `0x09` for both Tab and Ctrl+I (and `0x0D` for both Enter and Ctrl+M). go-tui maps these bytes to `KeyTab` and `KeyEnter`, so `KeyCtrlI` and `KeyCtrlM` only fire when the Kitty keyboard protocol is active and the terminal sends distinct sequences. If you need Tab or Enter handling without Kitty, bind `KeyTab` or `KeyEnter`.

## Modifier Flags

```go
type Modifier uint8

const (
    ModNone  Modifier = 0
    ModCtrl  Modifier = 1 << iota  // 1
    ModAlt                         // 2
    ModShift                       // 4
)
```

Modifiers are bit flags and can be combined with `|`:

```go
// Check for Ctrl+Shift
if ke.Mod == tui.ModCtrl|tui.ModShift {
    // ...
}
```

**Methods:**

### Has

```go
func (m Modifier) Has(mod Modifier) bool
```

Tests whether a specific modifier is present in the set.

```go
if ke.Mod.Has(tui.ModAlt) {
    // Alt was held
}
```

### String

```go
func (m Modifier) String() string
```

Returns a human-readable representation. Multiple modifiers are joined with `+`:

```go
tui.ModNone.String()                     // "None"
tui.ModCtrl.String()                     // "Ctrl"
(tui.ModCtrl | tui.ModShift).String()    // "Ctrl+Shift"
```

## MouseButton Constants

```go
type MouseButton int
```

| Constant | Description |
|----------|-------------|
| `MouseLeft` | Left (primary) button. |
| `MouseMiddle` | Middle button (scroll wheel click). |
| `MouseRight` | Right (secondary) button. |
| `MouseWheelUp` | Scroll wheel up. |
| `MouseWheelDown` | Scroll wheel down. |
| `MouseNone` | No button (motion events). |

## MouseAction Constants

```go
type MouseAction int
```

| Constant | Description |
|----------|-------------|
| `MousePress` | A button was pressed down. |
| `MouseRelease` | A button was released. |
| `MouseDrag` | Mouse moved while a button is held. |

## KeyMap

```go
type KeyMap []KeyBinding
```

A `KeyMap` is a slice of key bindings. Components return one from their `KeyMap()` method (the `KeyListener` interface). The framework collects bindings from all mounted components in tree order and checks them against each incoming key event. The first matching binding fires; if that binding has `Stop: true`, no further bindings run.

```go
func (a *myApp) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.On(tui.KeyEscape, func(ke tui.KeyEvent) {
            ke.App().Stop()
        }),
        tui.On(tui.Rune('q'), func(ke tui.KeyEvent) {
            ke.App().Stop()
        }),
    }
}
```

## KeyBinding

```go
type KeyBinding struct {
    Pattern KeyPattern
    Handler func(KeyEvent)
    Stop    bool
}
```

Associates a key pattern with a handler function. When `Stop` is `true`, the event does not propagate to any later bindings in the dispatch table.

## KeyPattern

```go
type KeyPattern struct {
    Key           Key
    Rune          rune
    AnyRune       bool
    Mod           Modifier
    ExcludeMods   Modifier
}
```

Describes which key events a binding matches.

| Field | Type | Description |
|-------|------|-------------|
| `Key` | `Key` | Match a specific special key (`KeyEscape`, `KeyEnter`, etc.), or `0` for none. |
| `Rune` | `rune` | Match a specific printable character, or `0` for none. |
| `AnyRune` | `bool` | When `true`, match any printable character. |
| `Mod` | `Modifier` | When non-zero, the event must have exactly these modifiers. |
| `ExcludeMods` | `Modifier` | Reject the event if any of these modifiers are present. |

You don't usually construct `KeyPattern` directly. Use the helper functions below instead.

## KeyMap Helper Functions

These functions build `KeyBinding` values for common use cases. They accept a `KeyMatcher` that describes which key events to match.

### KeyMatcher

A `KeyMatcher` describes which key events a binding should match. Three implementations are available:

- **Key constants** (`tui.KeyEscape`, `tui.KeyEnter`, etc.) match specific special keys directly.
- **`tui.Rune(r rune)`** returns a `RuneSpec` that matches a specific printable character.
- **`tui.AnyRune`** matches any printable character.

Both `Key` and `RuneSpec` support modifier methods that return a new matcher requiring the specified modifier:

```go
tui.KeyUp.Shift()       // Match Shift+Up
tui.KeyUp.Ctrl()        // Match Ctrl+Up
tui.KeyUp.Alt()         // Match Alt+Up
tui.Rune('s').Ctrl()    // Match Ctrl+S
tui.Rune('x').Alt()     // Match Alt+X
```

### On

```go
func On(m KeyMatcher, handler func(KeyEvent)) KeyBinding
```

Creates a binding that matches the given key pattern. The event continues propagating to later bindings.

```go
tui.On(tui.KeyEnter, func(ke tui.KeyEvent) {
    a.submit()
})

tui.On(tui.Rune('+'), func(ke tui.KeyEvent) {
    a.count.Update(func(v int) int { return v + 1 })
})

tui.On(tui.AnyRune, func(ke tui.KeyEvent) {
    a.buffer.Update(func(s string) string {
        return s + string(ke.Rune)
    })
})
```

### OnStop

```go
func OnStop(m KeyMatcher, handler func(KeyEvent)) KeyBinding
```

Same as `On`, but stops propagation after the handler runs. No later bindings will fire for this event.

```go
tui.OnStop(tui.KeyEnter, func(ke tui.KeyEvent) {
    a.submit() // this component owns Enter exclusively
})

tui.OnStop(tui.Rune('/'), func(ke tui.KeyEvent) {
    a.activateSearch() // capture '/' before anything else sees it
})

tui.OnStop(tui.AnyRune, func(ke tui.KeyEvent) {
    a.searchQuery.Update(func(s string) string {
        return s + string(ke.Rune)
    })
})
```

### OnFocused

```go
func OnFocused(m KeyMatcher, handler func(KeyEvent)) KeyBinding
```

Creates a binding that only fires when the component has focus. Stops propagation when it matches. Useful for focus-gated input handling.

```go
tui.OnFocused(tui.AnyRune, func(ke tui.KeyEvent) {
    a.textInput.Update(func(s string) string {
        return s + string(ke.Rune)
    })
})
```

### OnPreemptStop

```go
func OnPreemptStop(m KeyMatcher, handler func(KeyEvent)) KeyBinding
```

Creates a preemptive stop-propagation binding. Fires before all normal handlers in the dispatch table, preventing parent components from seeing the event. Used internally by Modal to block parent key handlers when the overlay is open.

```go
// Block all keys from reaching parent handlers
tui.OnPreemptStop(tui.AnyKey, func(ke tui.KeyEvent) {})

// Preemptive Escape handler
tui.OnPreemptStop(tui.KeyEscape, func(ke tui.KeyEvent) {
    closeOverlay()
})
```

## Component Interfaces

### KeyListener

```go
type KeyListener interface {
    KeyMap() KeyMap
}
```

Implement this on a struct component to handle keyboard input. The framework calls `KeyMap()` during each tree walk when the component is dirty, so you can return different bindings based on current state.

```go
func (a *myApp) KeyMap() tui.KeyMap {
    if a.searchActive.Get() {
        return tui.KeyMap{
            tui.OnStop(tui.KeyEscape, func(ke tui.KeyEvent) {
                a.searchActive.Set(false)
            }),
            tui.OnStop(tui.AnyRune, func(ke tui.KeyEvent) {
                a.searchQuery.Update(func(s string) string {
                    return s + string(ke.Rune)
                })
            }),
        }
    }
    return tui.KeyMap{
        tui.On(tui.KeyEscape, func(ke tui.KeyEvent) {
            ke.App().Stop()
        }),
        tui.On(tui.Rune('/'), func(ke tui.KeyEvent) {
            a.searchActive.Set(true)
        }),
    }
}
```

### MouseListener

```go
type MouseListener interface {
    HandleMouse(MouseEvent) bool
}
```

Implement this to handle mouse input. The framework walks the component tree and dispatches mouse events to each `MouseListener`. Return `true` if the event was consumed.

```go
func (a *myApp) HandleMouse(me tui.MouseEvent) bool {
    return tui.HandleClicks(me,
        tui.Click(a.saveBtn, a.save),
        tui.Click(a.cancelBtn, a.cancel),
    )
}
```

## Click Handling

For full details on ref-based click handling, see [Refs Reference](refs.md). The relevant functions:

### Click

```go
func Click(ref *Ref, fn func()) ClickBinding
```

Creates a binding between an element ref and a click handler.

### HandleClicks

```go
func HandleClicks(me MouseEvent, bindings ...ClickBinding) bool
```

Tests a mouse event against a list of click bindings. Only left-button press events are matched. Returns `true` if any binding's ref contained the click coordinates.

`HandleClicks` checks `MouseLeft` + `MousePress`. Other buttons and actions are ignored. The first binding whose ref element contains the click point `(X, Y)` fires, and the function returns `true`.

```go
func (c *counter) HandleMouse(me tui.MouseEvent) bool {
    return tui.HandleClicks(me,
        tui.Click(c.incrementBtn, c.increment),
        tui.Click(c.decrementBtn, c.decrement),
    )
}
```

## Event Dispatch Flow

Here's how events flow through the system.

**Keyboard events:**

1. Terminal input is read and parsed into a `KeyEvent`.
2. If the app uses the component model (struct components with `KeyMap()`), the dispatch table is built from all `KeyListener` components in tree order.
3. **Preemptive pass**: bindings marked as preemptive (e.g., modal catch-all) fire first. If any stops the event, normal dispatch is skipped entirely.
4. Bindings are checked in order. The first match fires. If `Stop` is true, dispatch ends.
5. If no binding stopped the event, it falls through to `App.Dispatch()` and the focus manager for element-level handlers.
6. In legacy mode (no components), `WithGlobalKeyHandler` runs first. If it returns `true`, the event is consumed.

**Mouse events:**

1. Terminal input is read and parsed into a `MouseEvent`.
2. The framework walks the component tree and dispatches to each `MouseListener`.
3. If any listener returns `true`, the event is consumed.
4. Otherwise, the event goes to `App.Dispatch()` which does hit-testing against the element tree.

**Resize events:**

1. The terminal reports a size change.
2. `App.Dispatch()` handles it directly: resizes the buffer, marks the root dirty, and schedules a full redraw.

## Global Key Handler

For apps not using the component model, you can set a global key handler that runs before focus-based dispatch.

### WithGlobalKeyHandler (AppOption)

```go
func WithGlobalKeyHandler(fn func(KeyEvent) bool) AppOption
```

Sets a handler that runs before key events reach the focus manager. Return `true` to consume the event.

```go
app, err := tui.NewApp(
    tui.WithRootComponent(MyApp()),
    tui.WithGlobalKeyHandler(func(ke tui.KeyEvent) bool {
        if ke.Key == tui.KeyCtrlC {
            // handle Ctrl+C globally
            return true
        }
        return false
    }),
)
```

### SetGlobalKeyHandler

```go
func (a *App) SetGlobalKeyHandler(fn func(KeyEvent) bool)
```

Sets or replaces the global key handler at runtime. Pass `nil` to remove it.

## See Also

- [State Reference](state.md) — reactive state that key handlers typically modify
- [Refs Reference](refs.md) — element references used for click hit-testing
- [Focus Reference](focus.md) — focus management and Tab navigation
- [App Reference](app.md) — app lifecycle, `WithMouse()`, `Dispatch()`
