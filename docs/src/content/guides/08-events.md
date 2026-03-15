# Event Handling

## Overview

go-tui routes keyboard and mouse events to your components through two interfaces: `KeyListener` for keyboard input and `MouseListener` for mouse clicks. The `KeyMap` system lets you declare key bindings as data. The framework collects bindings from all components in the tree, matches incoming key events against them, and calls the right handler. Mouse events work through refs, where you attach a `Ref` to an element and use `HandleClicks` to respond when the user clicks on it.

## KeyMap Basics

To handle keyboard input, implement the `KeyListener` interface on your struct component. It has one method, `KeyMap()`, which returns a `tui.KeyMap`. This is just a slice of `KeyBinding` values, each pairing a key pattern with a handler function:

```go
func (a *myApp) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.On(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
        tui.On(tui.Rune('+'), func(ke tui.KeyEvent) {
            a.count.Update(func(v int) int { return v + 1 })
        }),
        tui.On(tui.Rune('-'), func(ke tui.KeyEvent) {
            a.count.Update(func(v int) int { return v - 1 })
        }),
    }
}
```

The framework calls `KeyMap()` on every render cycle, so your bindings can change based on state. Bindings are checked in order within each component, and the first match wins within that component. If no binding has `Stop` set, the event continues to other components in the tree.

## Key Bindings

Three helper functions cover the binding patterns, each accepting a `KeyMatcher`:

| Function | Propagation | Use When |
|---|---|---|
| `On(matcher, handler)` | Continues | Default binding; other components can also handle the event |
| `OnStop(matcher, handler)` | Stops | Exclusive ownership; no other component sees the event |
| `OnFocused(matcher, handler)` | Stops (focus-gated) | Only fires when the component has focus |

A `KeyMatcher` can be a Key constant (`tui.KeyEscape`), a specific rune (`tui.Rune('q')`), or the catch-all `tui.AnyRune`. Key and RuneSpec both support `.Ctrl()`, `.Alt()`, and `.Shift()` modifier methods.

`On` with a Key constant matches special keys like Escape, Enter, and arrow keys. `On` with `Rune('x')` matches a specific printable character. `On` with `AnyRune` catches all printable characters, which is useful for text input.

The `OnStop` variant prevents other components from seeing the event after your handler runs. `On` lets the event continue through the component tree. More on propagation in a later section.

A quick example showing the difference between key and rune matching:

```go
func (a *myApp) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        // Match the Escape key (special key)
        tui.On(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),

        // Match the 'q' character (printable rune)
        tui.On(tui.Rune('q'), func(ke tui.KeyEvent) { ke.App().Stop() }),

        // Match any printable character
        tui.On(tui.AnyRune,func(ke tui.KeyEvent) {
            a.lastChar.Set(string(ke.Rune))
        }),
    }
}
```

## Special Keys

go-tui defines constants for every non-printable key the terminal can report:

**Navigation:**
`KeyUp`, `KeyDown`, `KeyLeft`, `KeyRight`, `KeyHome`, `KeyEnd`, `KeyPageUp`, `KeyPageDown`

**Editing:**
`KeyEnter`, `KeyTab`, `KeyBackspace`, `KeyDelete`, `KeyInsert`

**Control:**
`KeyEscape`, `KeyCtrlA` through `KeyCtrlZ`, `KeyCtrlSpace` (note: `KeyCtrlH`, `KeyCtrlI`, and `KeyCtrlM` are aliases for `KeyBackspace`, `KeyTab`, and `KeyEnter`; see Terminal Byte Aliases below)

**Function keys:**
`KeyF1` through `KeyF12`

Printable characters (letters, numbers, symbols) come through as `KeyRune` with the character in the `Rune` field. You generally don't check for `KeyRune` directly; the `Rune()` and `AnyRune` matchers handle that for you.

### Terminal Byte Aliases

Three Ctrl+letter combinations produce the same byte as a functional key:

| Alias | Same as | Terminal byte |
|---|---|---|
| `KeyCtrlH` | `KeyBackspace` | `0x08` |
| `KeyCtrlI` | `KeyTab` | `0x09` |
| `KeyCtrlM` | `KeyEnter` | `0x0D` |

In legacy mode, these are true aliases, not separate keys. `KeyCtrlH` and `KeyBackspace` are the same constant, so binding either one matches both. For example, `On(tui.KeyCtrlH, handler)` fires when the user presses Backspace, and `On(tui.KeyBackspace, handler)` fires when the user presses Ctrl+H.

If the terminal supports the Kitty keyboard protocol (negotiated automatically on startup), these keys become distinguishable: Backspace arrives as `KeyBackspace` while Ctrl+H arrives as `KeyEvent{Key: KeyRune, Rune: 'h', Mod: ModCtrl}`. Use `On(tui.Rune('h').Ctrl(), handler)` to handle Ctrl+H separately from Backspace when the Kitty protocol is active.

```go
// In legacy mode, these two bindings are identical:
tui.On(tui.KeyCtrlH, handler)    // matches Backspace AND Ctrl+H
tui.On(tui.KeyBackspace, handler) // matches Backspace AND Ctrl+H
```

Use whichever name best communicates your intent. If you're building a text editor where Backspace deletes a character, use `KeyBackspace`. If you're binding Ctrl+H to open a help panel, use `KeyCtrlH`, but keep in mind that in legacy mode Backspace will also trigger it.

## KeyEvent Properties

Every handler receives a `KeyEvent` with these fields and methods:

```go
ke.Key       // The Key constant (KeyEscape, KeyEnter, KeyRune, etc.)
ke.Rune      // The character, if Key == KeyRune (e.g., 'a', '+', '/')
ke.Mod       // Modifier flags: ModCtrl, ModAlt, ModShift
ke.IsRune()  // True when Key == KeyRune (a printable character)
ke.Is(key, mods...)  // Check key and optional modifiers in one call
ke.Char()    // Returns Rune if IsRune(), otherwise 0
ke.App()     // The running App instance
```

`ke.App()` gives you access to the application from inside any handler. The most common use is `ke.App().Stop()` to quit, but you can also call `ke.App().Batch()`, `ke.App().PrintAbove()`, or any other app method.

`ke.Is()` combines key and modifier checks:

```go
tui.On(tui.KeyCtrlS, func(ke tui.KeyEvent) {
    // Fires on Ctrl+S
    save()
}),
```

## Modifier Keys

The `Modifier` type is a bitmask with three flags:

| Constant | Description |
|---|---|
| `ModCtrl` | Control key held |
| `ModAlt` | Alt/Option key held |
| `ModShift` | Shift key held |

Check modifiers with `Has()`:

```go
if ke.Mod.Has(tui.ModAlt) {
    // Alt was held during this key event
}
```

Control keys have their own constants (`KeyCtrlA` through `KeyCtrlZ`), but you can also use the modifier method: `tui.Rune('s').Ctrl()` is equivalent to using `tui.KeyCtrlS` directly.

## Stop Propagation

When multiple components in the tree define key bindings, the framework walks them in breadth-first order, visiting shallower components before deeper ones. By default, every matching handler fires. The "Stop" variants change this: once a Stop handler matches, no further handlers for that key will run.

This matters when you have nested components. Consider a parent that uses `j`/`k` for navigation and a child search bar that needs all printable characters:

```go
// Parent: uses j/k for navigation (non-stop)
func (a *myApp) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.On(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
        tui.On(tui.Rune('j'), func(ke tui.KeyEvent) { a.selectNext() }),
        tui.On(tui.Rune('k'), func(ke tui.KeyEvent) { a.selectPrev() }),
        tui.On(tui.Rune('/'), func(ke tui.KeyEvent) { a.searchActive.Set(true) }),
    }
}
```

```go
// Child search bar: captures all runes when active (stop)
func (s *searchBar) KeyMap() tui.KeyMap {
    if !s.active.Get() {
        return nil
    }
    return tui.KeyMap{
        tui.OnStop(tui.AnyRune,s.appendChar),
        tui.OnStop(tui.KeyBackspace, s.deleteChar),
        tui.OnStop(tui.KeyEnter, s.submit),
        tui.OnStop(tui.KeyEscape, s.deactivate),
    }
}
```

When search is inactive, `KeyMap()` returns `nil` and the parent handles `j`/`k` normally. When search becomes active, the child's `OnStop(tui.AnyRune, ...)` grabs every printable character before the parent sees it. Pressing `j` types a "j" into the search bar instead of moving the selection.

This conditional KeyMap pattern is the standard way to build modal interfaces. The `KeyMap()` method runs on each render, so switching `active` to false immediately removes the child's bindings.

## Mouse Events

Mouse support requires an explicit opt-in when creating the app:

```go
app, err := tui.NewApp(
    tui.WithRootComponent(MyApp()),
    tui.WithMouse(),
)
```

To handle mouse events, implement the `MouseListener` interface:

```go
type MouseListener interface {
    HandleMouse(MouseEvent) bool
}
```

A `MouseEvent` has these fields:

```go
me.Button  // MouseLeft, MouseMiddle, MouseRight, MouseWheelUp, MouseWheelDown, MouseNone
me.Action  // MousePress, MouseRelease, MouseDrag
me.X, me.Y // 0-indexed column and row
me.Mod     // Modifier flags (Ctrl, Alt, Shift)
me.App()   // The running App instance
```

Return `true` from `HandleMouse` to indicate you handled the event. Return `false` to let other components try.

## Click Handling with Refs

Raw `X`/`Y` coordinates are cumbersome to work with directly. The `HandleClicks` helper does ref-based hit testing, so you can bind click handlers to specific elements without manual coordinate math.

The pattern has three steps:

**1. Create refs as struct fields:**

```go
type colorMixer struct {
    red   *tui.State[int]
    green *tui.State[int]

    redUpBtn   *tui.Ref
    redDnBtn   *tui.Ref
    greenUpBtn *tui.Ref
    greenDnBtn *tui.Ref
}

func ColorMixer() *colorMixer {
    return &colorMixer{
        red:        tui.NewState(128),
        green:      tui.NewState(64),
        redUpBtn:   tui.NewRef(),
        redDnBtn:   tui.NewRef(),
        greenUpBtn: tui.NewRef(),
        greenDnBtn: tui.NewRef(),
    }
}
```

**2. Attach refs to elements in your render method:**

```gsx
templ (c *colorMixer) Render() {
    <div class="flex gap-2 p-1">
        <div class="flex-col items-center gap-1">
            <span class="text-red font-bold">Red</span>
            <button ref={c.redUpBtn} class="px-2">{" + "}</button>
            <span class="text-red font-bold">{fmt.Sprintf("%d", c.red.Get())}</span>
            <button ref={c.redDnBtn} class="px-2">{" - "}</button>
        </div>
        <div class="flex-col items-center gap-1">
            <span class="text-green font-bold">Green</span>
            <button ref={c.greenUpBtn} class="px-2">{" + "}</button>
            <span class="text-green font-bold">{fmt.Sprintf("%d", c.green.Get())}</span>
            <button ref={c.greenDnBtn} class="px-2">{" - "}</button>
        </div>
    </div>
}
```

**3. Wire up HandleMouse with HandleClicks:**

```go
func (c *colorMixer) HandleMouse(me tui.MouseEvent) bool {
    return tui.HandleClicks(me,
        tui.Click(c.redUpBtn, func() { c.adjustRed(16) }),
        tui.Click(c.redDnBtn, func() { c.adjustRed(-16) }),
        tui.Click(c.greenUpBtn, func() { c.adjustGreen(16) }),
        tui.Click(c.greenDnBtn, func() { c.adjustGreen(-16) }),
    )
}
```

`HandleClicks` only responds to left-button presses. It checks each binding in order, calling the first handler whose ref element contains the click coordinates. It returns `true` if a click was handled.

## App-Level Key Handling

For keys that should be caught before any component sees them, use a global key handler. Set it as an app option or at runtime:

```go
// At creation
app, err := tui.NewApp(
    tui.WithRootComponent(MyApp()),
    tui.WithGlobalKeyHandler(func(ke tui.KeyEvent) bool {
        if ke.Key == tui.KeyCtrlC {
            // handle globally
            return true // consumed, components won't see it
        }
        return false // pass through to components
    }),
)

// At runtime
app.SetGlobalKeyHandler(func(ke tui.KeyEvent) bool {
    // ...
    return false
})
```

Return `true` to consume the event. Return `false` to let it continue to the component tree.

Most apps won't need this. The `KeyMap` system on components covers almost every case. Global handlers are for when you need to intercept keys regardless of which component has focus or what mode the app is in.

## Complete Example

This keyboard explorer tracks which keys have been pressed, using `On(tui.AnyRune, ...)` as a catch-all for printable characters and `On` with Key constants for special keys:

```gsx
package main

import (
    "fmt"
    tui "github.com/grindlemire/go-tui"
)

type explorer struct {
    lastKey  *tui.State[string]
    keyCount *tui.State[int]
}

func Explorer() *explorer {
    return &explorer{
        lastKey:  tui.NewState("(none)"),
        keyCount: tui.NewState(0),
    }
}

func (e *explorer) record(name string) {
    e.keyCount.Set(e.keyCount.Get() + 1)
    e.lastKey.Set(name)
}

func (e *explorer) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.OnStop(tui.Rune('q'), func(ke tui.KeyEvent) { ke.App().Stop() }),
        tui.On(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
        tui.On(tui.AnyRune,func(ke tui.KeyEvent) {
            e.record(fmt.Sprintf("'%c' (rune)", ke.Rune))
        }),
        // KeyCtrlH, KeyCtrlI, and KeyCtrlM are aliases for KeyBackspace,
        // KeyTab, and KeyEnter (same terminal byte), so you can use
        // either name here and both will match.
        tui.On(tui.KeyEnter, func(ke tui.KeyEvent) { e.record("Enter") }),
        tui.On(tui.KeyTab, func(ke tui.KeyEvent) { e.record("Tab") }),
        tui.On(tui.KeyBackspace, func(ke tui.KeyEvent) { e.record("Backspace") }),
        tui.On(tui.KeyUp, func(ke tui.KeyEvent) { e.record("Up") }),
        tui.On(tui.KeyDown, func(ke tui.KeyEvent) { e.record("Down") }),
        tui.On(tui.KeyLeft, func(ke tui.KeyEvent) { e.record("Left") }),
        tui.On(tui.KeyRight, func(ke tui.KeyEvent) { e.record("Right") }),
        tui.On(tui.KeyCtrlA, func(ke tui.KeyEvent) { e.record("Ctrl+A") }),
        tui.On(tui.KeyCtrlS, func(ke tui.KeyEvent) { e.record("Ctrl+S") }),
    }
}

templ (e *explorer) Render() {
    <div class="flex-col gap-1 p-2 border-rounded border-cyan">
        <span class="text-gradient-cyan-magenta font-bold">Keyboard Explorer</span>
        <hr class="border-single" />

        <div class="flex gap-2">
            <span class="font-dim">Last Key:</span>
            <span class="text-cyan font-bold">{e.lastKey.Get()}</span>
        </div>
        <div class="flex gap-2">
            <span class="font-dim">Key Count:</span>
            <span class="text-cyan font-bold">{fmt.Sprintf("%d", e.keyCount.Get())}</span>
        </div>

        <br />
        <span class="font-dim">Press any key to see it displayed above</span>
        <span class="font-dim">Press q or Esc to quit</span>
    </div>
}
```

With `main.go`:

```go
package main

import (
    "fmt"
    "os"

    tui "github.com/grindlemire/go-tui"
)

func main() {
    app, err := tui.NewApp(
        tui.WithRootComponent(Explorer()),
    )
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    defer app.Close()

    if err := app.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

Generate and run:

```bash
tui generate ./...
go run .
```

Here's what the event explorer looks like:

![Event Handling screenshot](/guides/08.png)

## Next Steps

- [Scrolling](scrolling) -- Scrollable containers for content that exceeds available space
- [Timers, Watchers, and Channels](watchers) -- Background operations with timers and Go channels
