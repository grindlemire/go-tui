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
| `*tui.RefList` | `tui.NewRefList()` | You have elements in a `for` loop, accessed by index |
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

## Refs in Loops

When you use `ref=` inside a `for` loop, the generator automatically uses the right ref type based on whether a `key` attribute is present.

### RefList (no key)

Without a `key`, the ref becomes a `RefList`, an ordered collection populated with `Append` on each iteration:

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
for _, item := range a.items {
    <span ref={a.itemRefs} class="p-1">{item}</span>
}
```

The generated code calls `a.itemRefs.Append(el)` for each iteration. Access elements with `.At(i)` or `.All()`.

### RefMap with key

Add a `key` attribute to turn the ref into a `RefMap[K]`. Each element is stored under its key:

```go
type tabApp struct {
    presetBtns *tui.RefMap[string]
}
```

```gsx
for _, p := range presets {
    <button ref={a.presetBtns} key={p.name} class="px-1">{p.name}</button>
}
```

The generated code calls `a.presetBtns.Put(p.name, el)` for each iteration. Look up elements with `.Get(key)` or iterate with `.All()`.

This is useful when you need to identify *which* element was clicked by its key rather than by position.

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
    // Check single-ref button clicks
    if tui.HandleClicks(me,
        tui.Click(c.redUpBtn, func() { c.adjustRed(16) }),
        tui.Click(c.redDnBtn, func() { c.adjustRed(-16) }),
        tui.Click(c.greenUpBtn, func() { c.adjustGreen(16) }),
        tui.Click(c.greenDnBtn, func() { c.adjustGreen(-16) }),
        tui.Click(c.blueUpBtn, func() { c.adjustBlue(16) }),
        tui.Click(c.blueDnBtn, func() { c.adjustBlue(-16) }),
    ) {
        return true
    }

    // Check keyed-ref preset button clicks via RefMap
    if me.Button == tui.MouseLeft && me.Action == tui.MousePress {
        for name, el := range c.presetBtns.All() {
            if el != nil && el.ContainsPoint(me.X, me.Y) {
                c.applyPreset(name)
                return true
            }
        }
    }

    return false
}
```

Both input methods call the same `adjust*` methods, so the behavior stays consistent. The preset buttons use `RefMap.All()` to iterate and hit-test by key, letting us identify which preset was clicked by name.

## Complete Example

This color mixer lets you adjust RGB values with both keyboard shortcuts and mouse clicks. Each color channel has a visual bar, a value readout, and +/- buttons. It also includes clickable preset color buttons using `RefMap` with `key`:

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

    redUpBtn     *tui.Ref
    redDnBtn     *tui.Ref
    greenUpBtn   *tui.Ref
    greenDnBtn   *tui.Ref
    blueUpBtn    *tui.Ref
    blueDnBtn    *tui.Ref
    presetBtns   *tui.RefMap[string]
    activePreset *tui.State[string]
}

func ColorMixer() *colorMixer {
    return &colorMixer{
        red:          tui.NewState(128),
        green:        tui.NewState(64),
        blue:         tui.NewState(200),
        redUpBtn:     tui.NewRef(),
        redDnBtn:     tui.NewRef(),
        greenUpBtn:   tui.NewRef(),
        greenDnBtn:   tui.NewRef(),
        blueUpBtn:    tui.NewRef(),
        blueDnBtn:    tui.NewRef(),
        presetBtns:   tui.NewRefMap[string](),
        activePreset: tui.NewState(""),
    }
}

type preset struct {
    name    string
    r, g, b int
}

var presets = []preset{
    {"Sunset", 255, 128, 0},
    {"Ocean", 0, 100, 255},
    {"Forest", 34, 180, 34},
    {"Rose", 255, 64, 128},
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
    c.activePreset.Set("")
}

func (c *colorMixer) adjustGreen(delta int) {
    c.green.Set(clamp(c.green.Get()+delta, 0, 255))
    c.activePreset.Set("")
}

func (c *colorMixer) adjustBlue(delta int) {
    c.blue.Set(clamp(c.blue.Get()+delta, 0, 255))
    c.activePreset.Set("")
}

func (c *colorMixer) applyPreset(name string) {
    for _, p := range presets {
        if p.name == name {
            c.red.Set(p.r)
            c.green.Set(p.g)
            c.blue.Set(p.b)
            c.activePreset.Set(name)
            return
        }
    }
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
    // Check single-ref button clicks
    if tui.HandleClicks(me,
        tui.Click(c.redUpBtn, func() { c.adjustRed(16) }),
        tui.Click(c.redDnBtn, func() { c.adjustRed(-16) }),
        tui.Click(c.greenUpBtn, func() { c.adjustGreen(16) }),
        tui.Click(c.greenDnBtn, func() { c.adjustGreen(-16) }),
        tui.Click(c.blueUpBtn, func() { c.adjustBlue(16) }),
        tui.Click(c.blueDnBtn, func() { c.adjustBlue(-16) }),
    ) {
        return true
    }

    // Check keyed-ref preset button clicks via RefMap
    if me.Button == tui.MouseLeft && me.Action == tui.MousePress {
        for name, el := range c.presetBtns.All() {
            if el != nil && el.ContainsPoint(me.X, me.Y) {
                c.applyPreset(name)
                return true
            }
        }
    }

    return false
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
    <div class="flex-col p-1 border-rounded border-cyan">
        <span class="text-gradient-cyan-magenta font-bold">Color Mixer</span>

        // Color preview
        <div class="flex-col items-center border-rounded p-1">
            <span class="text-gradient-cyan-magenta font-bold">Preview</span>
            <div backgroundGradient={tui.NewGradient(tui.Black, tui.RGBColor(uint8(c.red.Get()), uint8(c.green.Get()), uint8(c.blue.Get())))} height={2} width={30}>
                <span>{" "}</span>
            </div>
            <div class="flex gap-2 justify-center">
                <span class="text-red font-bold">{fmt.Sprintf("R: %d", c.red.Get())}</span>
                <span class="text-green font-bold">{fmt.Sprintf("G: %d", c.green.Get())}</span>
                <span class="text-blue font-bold">{fmt.Sprintf("B: %d", c.blue.Get())}</span>
            </div>
        </div>

        // Color bars
        <div class="flex-col border-rounded p-1">
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
        <div class="flex gap-1">
            <div class="flex-col border-rounded p-1 items-center" flexGrow={1.0}>
                <span class="font-bold text-red">Red</span>
                <div class="flex gap-1 items-center">
                    <button ref={c.redDnBtn} class="px-1">{"-"}</button>
                    <span class="font-bold text-red">{fmt.Sprintf("%3d", c.red.Get())}</span>
                    <button ref={c.redUpBtn} class="px-1">{"+"}</button>
                </div>
            </div>
            <div class="flex-col border-rounded p-1 items-center" flexGrow={1.0}>
                <span class="font-bold text-green">Green</span>
                <div class="flex gap-1 items-center">
                    <button ref={c.greenDnBtn} class="px-1">{"-"}</button>
                    <span class="font-bold text-green">{fmt.Sprintf("%3d", c.green.Get())}</span>
                    <button ref={c.greenUpBtn} class="px-1">{"+"}</button>
                </div>
            </div>
            <div class="flex-col border-rounded p-1 items-center" flexGrow={1.0}>
                <span class="font-bold text-blue">Blue</span>
                <div class="flex gap-1 items-center">
                    <button ref={c.blueDnBtn} class="px-1">{"-"}</button>
                    <span class="font-bold text-blue">{fmt.Sprintf("%3d", c.blue.Get())}</span>
                    <button ref={c.blueUpBtn} class="px-1">{"+"}</button>
                </div>
            </div>
        </div>

        // Preset colors using RefMap with key
        <div class="flex gap-1 border-rounded p-1 items-center">
            <span class="font-bold">Presets:</span>
            for _, p := range presets {
                if p.name == c.activePreset.Get() {
                    <button ref={c.presetBtns} key={p.name} class="px-1 font-bold text-cyan">{p.name}</button>
                } else {
                    <button ref={c.presetBtns} key={p.name} class="px-1 font-dim">{p.name}</button>
                }
            }
        </div>

        <div class="flex justify-center">
            <span class="font-dim">r/g/b increase | R/G/B decrease | click buttons/presets | q quit</span>
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

Click the +/- buttons, use r/g/b keys to adjust colors, or click a preset to apply it:

![Refs and Click Handling screenshot](/guides/09.png)

## Next Steps

- [Built-in Elements](elements) - Reference guide to every HTML-like element in go-tui
- [Streaming Data](streaming) - Build a live data viewer with channels and auto-scroll
