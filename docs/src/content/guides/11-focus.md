# Focus Management

## Overview

Focus determines which element receives keyboard input. go-tui handles element-level focus automatically. Mark elements as focusable and use `app.FocusNext()` / `app.FocusPrev()` to cycle between them. For section-level switching (e.g., sidebar vs content panel), `FocusGroup` manages mutual-exclusion state using `*State[bool]` values with built-in Tab/Shift+Tab bindings.

## Making Elements Focusable

Mark an element as focusable with the `focusable` attribute in your `.gsx` template:

```gsx
<button focusable={true} class="px-2 border-single">Click me</button>
```

You can also attach focus and blur callbacks. These fire when the element gains or loses focus:

```gsx
<div
    focusable={true}
    onFocus={func(el *tui.Element) { el.SetBorderStyle(tui.NewStyle().Foreground(tui.ANSIColor(tui.Cyan))) }}
    onBlur={func(el *tui.Element) { el.SetBorderStyle(tui.NewStyle()) }}
    class="border-single p-1"
>
    <span>Focus me</span>
</div>
```

Setting `onFocus` or `onBlur` implicitly makes the element focusable, so you don't need to also set `focusable={true}`.

In Go code, the same behavior is available through option functions:

```go
elem := tui.New(
    tui.WithFocusable(true),
    tui.WithOnFocus(func(el *tui.Element) {
        el.SetBorderStyle(tui.NewStyle().Foreground(tui.ANSIColor(tui.Cyan)))
    }),
    tui.WithOnBlur(func(el *tui.Element) {
        el.SetBorderStyle(tui.NewStyle())
    }),
)
```

To check whether an element has focus, call `IsFocused()`:

```go
if s.myRef.El() != nil && s.myRef.El().IsFocused() {
    <span class="text-cyan">Focused</span>
} else {
    <span class="text-dim">Not focused</span>
}
```

## Focus Navigation

The `App` provides three methods for focus control:

- `app.FocusNext()` -- move focus to the next focusable element
- `app.FocusPrev()` -- move focus to the previous focusable element
- `app.Focused()` -- get the currently focused `Focusable`, or nil

Focus order follows document order (depth-first traversal of the element tree). `FocusNext` wraps from the last element back to the first; `FocusPrev` wraps from the first to the last.

Wire these up in your `KeyMap` to give users Tab/Shift+Tab navigation:

```go
func (f *myForm) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
        tui.OnKey(tui.KeyTab, func(ke tui.KeyEvent) { ke.App().FocusNext() }),
        // Note: Shift+Tab arrives as KeyTab with ModShift
        {
            Pattern: tui.KeyPattern{Key: tui.KeyTab, Mod: tui.ModShift},
            Handler: func(ke tui.KeyEvent) { ke.App().FocusPrev() },
        },
    }
}
```

When the app starts, the first focusable element in the tree is automatically focused. If no elements are focusable, nothing receives focus.

## FocusGroup

`FocusGroup` manages Tab/Shift+Tab cycling between logical sections of your UI. Instead of tracking individual elements, it works with `*State[bool]` values. Each member state is `true` when that section is active and `false` otherwise. The group maintains mutual exclusion: exactly one member is active at a time.

Create a focus group with two or more `*State[bool]` members:

```go
type myForm struct {
    sidebarActive *tui.State[bool]
    contentActive *tui.State[bool]
    footerActive  *tui.State[bool]
    focus         *tui.FocusGroup
}

func MyForm() *myForm {
    sidebar := tui.NewState(true)  // starts active
    content := tui.NewState(false)
    footer  := tui.NewState(false)

    return &myForm{
        sidebarActive: sidebar,
        contentActive: content,
        footerActive:  footer,
        focus:         tui.MustNewFocusGroup(sidebar, content, footer),
    }
}
```

`NewFocusGroup` sets the first member to `true` and all others to `false`. It returns an error if you pass fewer than 2 members. `MustNewFocusGroup` panics on error, which is fine for constructors where the member count is known at compile time.

### Navigation

`FocusGroup` has three methods:

- `fg.Next()` -- deactivate the current member, activate the next (wraps around)
- `fg.Prev()` -- deactivate the current member, activate the previous (wraps around)
- `fg.Current()` -- return the index of the currently active member

### Built-in KeyMap

`FocusGroup` implements the `KeyListener` interface. Its `KeyMap()` returns bindings for Tab (calls `Next`) and Shift+Tab (calls `Prev`). Include it in your component's `KeyMap` by spreading the group's bindings:

```go
func (f *myForm) KeyMap() tui.KeyMap {
    return append(f.focus.KeyMap(), []tui.KeyBinding{
        tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
        // section-specific bindings...
    }...)
}
```

### Using Active State in Render

Read each member state in your template to change styles based on which section is active:

```gsx
templ (f *myForm) Render() {
    <div class="flex gap-1 h-full">
        if f.sidebarActive.Get() {
            <div class="flex-col border-rounded border-cyan p-1" width={20}>
                <span class="text-cyan font-bold">Sidebar</span>
                <span>Use Tab to switch</span>
            </div>
        } else {
            <div class="flex-col border-rounded border-black p-1" width={20}>
                <span class="font-dim">Sidebar</span>
            </div>
        }

        if f.contentActive.Get() {
            <div class="flex-col grow border-rounded border-cyan p-1">
                <span class="text-cyan font-bold">Content</span>
                <span>This panel is active</span>
            </div>
        } else {
            <div class="flex-col grow border-rounded border-black p-1">
                <span class="font-dim">Content</span>
            </div>
        }

        if f.footerActive.Get() {
            <div class="flex-col border-rounded border-cyan p-1" width={20}>
                <span class="text-cyan font-bold">Footer</span>
            </div>
        } else {
            <div class="flex-col border-rounded border-black p-1" width={20}>
                <span class="font-dim">Footer</span>
            </div>
        }
    </div>
}
```

## Programmatic Focus

Elements provide direct focus control through these methods:

```go
element.Focus()              // give this element focus
element.Blur()               // remove focus from this element
element.IsFocused() bool     // check if this element has focus
element.IsFocusable() bool   // check if this element can receive focus
element.SetFocusable(bool)   // enable or disable focusability at runtime
```

`Focus()` and `Blur()` trigger the `onFocus` and `onBlur` callbacks if they were set via `WithOnFocus`/`WithOnBlur` or `SetOnFocus`/`SetOnBlur`.

You can also set callbacks programmatically after creation:

```go
element.SetOnFocus(func(el *tui.Element) {
    // called when element gains focus
})
element.SetOnBlur(func(el *tui.Element) {
    // called when element loses focus
})
```

Both `SetOnFocus` and `SetOnBlur` implicitly set the element as focusable.

## Complete Example

This form has three panels that highlight when active. Tab and Shift+Tab cycle between them using `FocusGroup`:

```gsx
package main

import (
    "fmt"
    tui "github.com/grindlemire/go-tui"
)

type panelForm struct {
    panels     []string
    panel1     *tui.State[bool]
    panel2     *tui.State[bool]
    panel3     *tui.State[bool]
    focus      *tui.FocusGroup
    clickCount *tui.State[int]
}

func PanelForm() *panelForm {
    p1 := tui.NewState(true)
    p2 := tui.NewState(false)
    p3 := tui.NewState(false)

    return &panelForm{
        panels:     []string{"Inbox", "Drafts", "Sent"},
        panel1:     p1,
        panel2:     p2,
        panel3:     p3,
        focus:      tui.MustNewFocusGroup(p1, p2, p3),
        clickCount: tui.NewState(0),
    }
}

func (p *panelForm) KeyMap() tui.KeyMap {
    return append(p.focus.KeyMap(), []tui.KeyBinding{
        tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
        tui.OnRune(' ', func(ke tui.KeyEvent) {
            p.clickCount.Update(func(v int) int { return v + 1 })
        }),
    }...)
}

templ (p *panelForm) Render() {
    <div class="flex-col gap-1 p-1">
        <span class="font-bold text-gradient-cyan-magenta">Focus Demo — Tab to switch, Space to interact</span>
        <div class="flex gap-1">
            for i, name := range p.panels {
                if i == p.focus.Current() {
                    <div class="flex-col border-rounded border-cyan p-1" width={20}>
                        <span class="text-cyan font-bold">{name}</span>
                        <span class="text-bright-white">{fmt.Sprintf("Actions: %d", p.clickCount.Get())}</span>
                    </div>
                } else {
                    <div class="flex-col border-rounded border-black p-1" width={20}>
                        <span class="font-dim">{name}</span>
                    </div>
                }
            }
        </div>
        <span class="text-dim">Press Esc to quit</span>
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
        tui.WithRootComponent(PanelForm()),
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

Tab between the panels to see focus in action:

![Focus Management screenshot](/guides/11.png)

## Next Steps

- [Building a Dashboard](dashboard) -- Build a live metrics dashboard from scratch
- [Events Guide](events) -- Keyboard and mouse event handling
