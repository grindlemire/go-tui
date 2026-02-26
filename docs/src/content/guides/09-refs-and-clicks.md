# Refs and Click Handling

## Overview

Refs let you hold a reference to a rendered element so you can interact with it from Go code. The most common use is mouse click handling: attach a ref to a button, then check if a click landed on that button. go-tui provides three ref types (`Ref`, `RefList`, and `RefMap[K]`) for single elements, indexed collections, and keyed collections.

## What is a Ref

A `Ref` is a pointer to a single element in the rendered tree. Create one with `tui.NewRef()` and attach it to an element using the `ref` attribute:

```go
type myApp struct {
    saveBtn *tui.Ref
}

func MyApp() *myApp {
    return &myApp{
        saveBtn: tui.NewRef(),
    }
}
```

```gsx
templ (a *myApp) Render() {
    <button ref={a.saveBtn} class="px-2">Save</button>
}
```

After the first render, `a.saveBtn.El()` returns the `*tui.Element` for that button. Before the first render it returns `nil`, so always check:

```go
if el := a.saveBtn.El(); el != nil {
    // safe to use el
}
```

The generated code calls `.Set()` on the ref each render cycle, so the ref always points to the current element instance.

## Ref Types

go-tui provides three ref types for different use cases:

| Type | Constructor | Use When |
|------|------------|----------|
| `*tui.Ref` | `tui.NewRef()` | You have a single element to reference |
| `*tui.RefList` | `tui.NewRefList()` | You have elements in a `@for` loop, accessed by index |
| `*tui.RefMap[K]` | `tui.NewRefMap[string]()` | You have elements keyed by a value (string, int, etc.) |

`Ref` stores one element. `RefList` stores elements by their loop index; use `.At(i)` to bind in the template and `.El(i)` to read back. `RefMap[K]` stores elements by an arbitrary key; use `.At(key)` to bind and `.El(key)` to read back.

## Click Handling Pattern

Mouse click handling follows a three-step pattern:

**1. Create refs** in your constructor:

```go
func MyApp() *myApp {
    return &myApp{
        saveBtn:   tui.NewRef(),
        cancelBtn: tui.NewRef(),
    }
}
```

**2. Bind refs** in your template:

```gsx
<button ref={a.saveBtn} class="px-2">Save</button>
<button ref={a.cancelBtn} class="px-2">Cancel</button>
```

**3. Wire up HandleMouse** with `HandleClicks`:

```go
func (a *myApp) HandleMouse(me tui.MouseEvent) bool {
    return tui.HandleClicks(me,
        tui.Click(a.saveBtn, a.onSave),
        tui.Click(a.cancelBtn, a.onCancel),
    )
}

func (a *myApp) onSave()   { /* ... */ }
func (a *myApp) onCancel() { /* ... */ }
```

`HandleClicks` checks each `Click` binding in order. When the mouse event is a left-click press that lands within a ref's element bounds, it calls that binding's handler and returns `true`. If no binding matches, it returns `false`.

## HandleClicks Details

`tui.HandleClicks` only responds to left-click press events (`MouseLeft` with `MousePress`). It does not fire on release, drag, or right-click. The bindings are checked in order, and the first match wins.

Each `tui.Click` binding takes a ref (any of the three types) and a `func()` handler. The framework checks whether the click coordinates fall within the element's rendered bounds. For `RefList` and `RefMap`, it checks all stored elements.

The `HandleMouse` method on your component implements the `MouseListener` interface. Return `true` if you handled the event, `false` to let it propagate.

## RefList for Loop Elements

When you render elements in a `@for` loop, use `RefList` to reference them by index:

```go
type listApp struct {
    items    []string
    itemRefs *tui.RefList
}

func ListApp() *listApp {
    return &listApp{
        items:    []string{"Alpha", "Beta", "Gamma"},
        itemRefs: tui.NewRefList(),
    }
}
```

```gsx
@for i, item := range a.items {
    <button ref={a.itemRefs.At(i)} class="px-2">{item}</button>
}
```

In `HandleMouse`, pass the `RefList` directly to `Click`. It checks all indexed elements:

```go
func (a *listApp) HandleMouse(me tui.MouseEvent) bool {
    return tui.HandleClicks(me,
        tui.Click(a.itemRefs, a.onItemClick),
    )
}
```

For keyed collections, `RefMap[K]` works the same way but uses `.At(key)` instead of `.At(index)`.

## Combining Keyboard and Mouse

The color mixer example wires both keyboard shortcuts and clickable buttons to the same actions:

```go
func (c *colorMixer) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
        tui.OnRune('r', func(ke tui.KeyEvent) { c.adjustRed(16) }),
        tui.OnRune('R', func(ke tui.KeyEvent) { c.adjustRed(-16) }),
        tui.OnRune('g', func(ke tui.KeyEvent) { c.adjustGreen(16) }),
        tui.OnRune('G', func(ke tui.KeyEvent) { c.adjustGreen(-16) }),
        tui.OnRune('b', func(ke tui.KeyEvent) { c.adjustBlue(16) }),
        tui.OnRune('B', func(ke tui.KeyEvent) { c.adjustBlue(-16) }),
    }
}

func (c *colorMixer) HandleMouse(me tui.MouseEvent) bool {
    return tui.HandleClicks(me,
        tui.Click(c.redUpBtn, func() { c.adjustRed(16) }),
        tui.Click(c.redDnBtn, func() { c.adjustRed(-16) }),
        tui.Click(c.greenUpBtn, func() { c.adjustGreen(16) }),
        tui.Click(c.greenDnBtn, func() { c.adjustGreen(-16) }),
        tui.Click(c.blueUpBtn, func() { c.adjustBlue(16) }),
        tui.Click(c.blueDnBtn, func() { c.adjustBlue(-16) }),
    )
}
```

Both input methods call the same `adjust*` methods, so the behavior stays consistent.

## Complete Example

This color mixer lets you adjust RGB values with both keyboard shortcuts and mouse clicks. Each color channel has a visual bar, a value readout, and +/- buttons:

```gsx
package main

import (
    "fmt"
    tui "github.com/grindlemire/go-tui"
)

type colorMixer struct {
    red   *tui.State[int]
    green *tui.State[int]
    blue  *tui.State[int]

    redUpBtn   *tui.Ref
    redDnBtn   *tui.Ref
    greenUpBtn *tui.Ref
    greenDnBtn *tui.Ref
    blueUpBtn  *tui.Ref
    blueDnBtn  *tui.Ref
}

func ColorMixer() *colorMixer {
    return &colorMixer{
        red:        tui.NewState(128),
        green:      tui.NewState(64),
        blue:       tui.NewState(200),
        redUpBtn:   tui.NewRef(),
        redDnBtn:   tui.NewRef(),
        greenUpBtn: tui.NewRef(),
        greenDnBtn: tui.NewRef(),
        blueUpBtn:  tui.NewRef(),
        blueDnBtn:  tui.NewRef(),
    }
}

func clamp(v, min, max int) int {
    if v < min {
        return min
    }
    if v > max {
        return max
    }
    return v
}

func (c *colorMixer) adjustRed(delta int) {
    c.red.Set(clamp(c.red.Get()+delta, 0, 255))
}

func (c *colorMixer) adjustGreen(delta int) {
    c.green.Set(clamp(c.green.Get()+delta, 0, 255))
}

func (c *colorMixer) adjustBlue(delta int) {
    c.blue.Set(clamp(c.blue.Get()+delta, 0, 255))
}

func (c *colorMixer) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
        tui.OnRune('q', func(ke tui.KeyEvent) { ke.App().Stop() }),
        tui.OnRune('r', func(ke tui.KeyEvent) { c.adjustRed(16) }),
        tui.OnRune('R', func(ke tui.KeyEvent) { c.adjustRed(-16) }),
        tui.OnRune('g', func(ke tui.KeyEvent) { c.adjustGreen(16) }),
        tui.OnRune('G', func(ke tui.KeyEvent) { c.adjustGreen(-16) }),
        tui.OnRune('b', func(ke tui.KeyEvent) { c.adjustBlue(16) }),
        tui.OnRune('B', func(ke tui.KeyEvent) { c.adjustBlue(-16) }),
    }
}

func (c *colorMixer) HandleMouse(me tui.MouseEvent) bool {
    return tui.HandleClicks(me,
        tui.Click(c.redUpBtn, func() { c.adjustRed(16) }),
        tui.Click(c.redDnBtn, func() { c.adjustRed(-16) }),
        tui.Click(c.greenUpBtn, func() { c.adjustGreen(16) }),
        tui.Click(c.greenDnBtn, func() { c.adjustGreen(-16) }),
        tui.Click(c.blueUpBtn, func() { c.adjustBlue(16) }),
        tui.Click(c.blueDnBtn, func() { c.adjustBlue(-16) }),
    )
}

func colorBar(value int) string {
    filled := value * 20 / 255
    bar := ""
    for i := 0; i < 20; i++ {
        if i < filled {
            bar += "█"
        } else {
            bar += "░"
        }
    }
    return bar
}

templ (c *colorMixer) Render() {
    <div class="flex-col p-2 gap-2 border-rounded border-cyan">
        <span class="text-gradient-cyan-magenta font-bold">Color Mixer</span>

        // Color preview
        <div class="flex-col items-center gap-1 border-rounded p-1">
            <span class="text-gradient-cyan-magenta font-bold">Preview</span>
            <div class="bg-gradient-cyan-magenta" height={3} width={30}>
                <span>{" "}</span>
            </div>
            <div class="flex gap-2 justify-center">
                <span class="text-red font-bold">{fmt.Sprintf("R: %d", c.red.Get())}</span>
                <span class="text-green font-bold">{fmt.Sprintf("G: %d", c.green.Get())}</span>
                <span class="text-blue font-bold">{fmt.Sprintf("B: %d", c.blue.Get())}</span>
            </div>
        </div>

        // Color bars
        <div class="flex-col gap-1 border-rounded p-1">
            <div class="flex gap-1">
                <span class="text-red font-bold w-5">Red</span>
                <span class="text-red">{colorBar(c.red.Get())}</span>
                <span class="text-red font-bold">{fmt.Sprintf("%3d", c.red.Get())}</span>
            </div>
            <div class="flex gap-1">
                <span class="text-green font-bold w-5">Grn</span>
                <span class="text-green">{colorBar(c.green.Get())}</span>
                <span class="text-green font-bold">{fmt.Sprintf("%3d", c.green.Get())}</span>
            </div>
            <div class="flex gap-1">
                <span class="text-blue font-bold w-5">Blu</span>
                <span class="text-blue">{colorBar(c.blue.Get())}</span>
                <span class="text-blue font-bold">{fmt.Sprintf("%3d", c.blue.Get())}</span>
            </div>
        </div>

        // Channel controls with refs
        <div class="flex gap-2">
            <div class="flex-col border-rounded p-1 gap-1 items-center" flexGrow={1.0}>
                <span class="font-bold text-red">Red</span>
                <button ref={c.redUpBtn} class="px-2">{" + "}</button>
                <span class="font-bold text-red">{fmt.Sprintf("%d", c.red.Get())}</span>
                <button ref={c.redDnBtn} class="px-2">{" - "}</button>
            </div>
            <div class="flex-col border-rounded p-1 gap-1 items-center" flexGrow={1.0}>
                <span class="font-bold text-green">Green</span>
                <button ref={c.greenUpBtn} class="px-2">{" + "}</button>
                <span class="font-bold text-green">{fmt.Sprintf("%d", c.green.Get())}</span>
                <button ref={c.greenDnBtn} class="px-2">{" - "}</button>
            </div>
            <div class="flex-col border-rounded p-1 gap-1 items-center" flexGrow={1.0}>
                <span class="font-bold text-blue">Blue</span>
                <button ref={c.blueUpBtn} class="px-2">{" + "}</button>
                <span class="font-bold text-blue">{fmt.Sprintf("%d", c.blue.Get())}</span>
                <button ref={c.blueDnBtn} class="px-2">{" - "}</button>
            </div>
        </div>

        <div class="flex justify-center">
            <span class="font-dim">r/g/b increase | R/G/B decrease | click buttons | q quit</span>
        </div>
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
        tui.WithRootComponent(ColorMixer()),
    )
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to create app: %v\n", err)
        os.Exit(1)
    }
    defer app.Close()

    if err := app.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "App error: %v\n", err)
        os.Exit(1)
    }
}
```

Generate and run:

```bash
tui generate ./...
go run .
```

Click the +/- buttons or use r/g/b keys to adjust colors:

![Refs and Click Handling screenshot](/guides/09.png)

## Next Steps

- [Built-in Elements](elements) - Reference guide to every HTML-like element in go-tui
- [Streaming Data](streaming) - Build a live data viewer with channels and auto-scroll
