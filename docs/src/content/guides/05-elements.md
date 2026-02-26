# Built-in Elements

## Overview

go-tui provides HTML-like elements (`<div>`, `<span>`, `<p>`, `<ul>`, `<li>`, `<button>`, `<input>`, `<table>`, `<progress>`, `<hr>`, `<br>`) that compile to `tui.New()` calls with appropriate defaults. Here's what each one does and when to use it.

## Container Elements

`<div>` is the primary layout container. It renders as a flexbox container with `Row` direction by default. Use Tailwind classes or attributes to configure direction, alignment, gaps, and sizing:

```gsx
// Horizontal layout (default)
<div class="flex gap-2">
    <span>Left</span>
    <span>Right</span>
</div>

// Vertical layout
<div class="flex-col gap-1 p-1 border-rounded">
    <span>Top</span>
    <span>Bottom</span>
</div>

// Nested layout
<div class="flex gap-1">
    <div class="flex-col border-single p-1" width={20}>
        <span>Sidebar</span>
    </div>
    <div class="flex-col grow border-single p-1">
        <span>Content</span>
    </div>
</div>
```

Every visible element in go-tui is ultimately a `<div>` with different default options applied.

## Text Elements

### span

`<span>` displays inline text. Use it for styled text content within a layout:

```gsx
<span>Plain text</span>
<span class="text-cyan font-bold">Styled text</span>
<span class="text-[#ff6600]">Hex-colored text</span>
```

### p

`<p>` renders paragraph text that wraps automatically when it exceeds the available width:

```gsx
<p>{"This paragraph text wraps automatically when the content exceeds the available width. Use <p> for longer text blocks."}</p>
```

Use `<span>` for short inline labels and `<p>` for longer text that should word-wrap.

## Separator Elements

### hr

`<hr>` draws a horizontal rule across the container width. It's self-closing:

```gsx
<div class="flex-col gap-1">
    <span>Above the line</span>
    <hr />
    <span>Below the line</span>
</div>
```

### br

`<br>` inserts a blank line break. Also self-closing:

```gsx
<div class="flex-col">
    <span>Line one</span>
    <br />
    <span>Line three (with a blank line above)</span>
</div>
```

## List Elements

`<ul>` creates a list container and `<li>` renders list items with bullet markers. Nest them together for bulleted lists:

```gsx
<ul class="flex-col p-1">
    <li><span>First item</span></li>
    <li><span>Second item</span></li>
    <li><span class="text-cyan">Third (styled)</span></li>
</ul>
```

Each `<li>` automatically prepends a bullet character. Put any content inside the `<li>`, typically a `<span>` with text, but it can contain other elements too.

## Table Element

`<table>` acts as a flex container for tabular data. Build tables by composing `<div>` rows with fixed-width columns and an `<hr>` separator between the header and body:

```gsx
<table class="flex-col p-1">
    // Header row
    <div class="flex gap-2">
        <span class="w-10 font-bold">Name</span>
        <span class="w-10 font-bold">Role</span>
        <span class="w-5 font-bold">Lvl</span>
    </div>
    <hr />
    // Data rows
    <div class="flex gap-2">
        <span class="w-10 text-cyan">Alice</span>
        <span class="w-10">Engineer</span>
        <span class="w-5 text-green">Sr</span>
    </div>
    <div class="flex gap-2">
        <span class="w-10 text-cyan">Bob</span>
        <span class="w-10">Designer</span>
        <span class="w-5 text-yellow">Jr</span>
    </div>
</table>
```

The fixed widths on each column (`w-10`, `w-5`) keep columns aligned across rows. Use `gap-2` on the row `<div>` for spacing between columns.

## Button Element

`<button>` renders a clickable button. Combine it with refs for mouse handling (see the [Refs and Clicks guide](refs-and-clicks)):

```gsx
<div class="flex gap-2">
    <button>{"Save"}</button>
    <button class="font-bold">{"Submit"}</button>
    <button disabled={true}>{"Disabled"}</button>
</div>
```

The `disabled` attribute visually dims the button. Wire up click handling through the `MouseListener` interface with `HandleClicks` and a `Ref` bound to the button.

## Input Element

`<input>` renders a text input field with `value` and `placeholder` attributes:

```gsx
<input value={s.inputValue.Get()} placeholder="Type here..." />
```

Bind the `value` attribute to a `State[string]` so the input updates reactively. The `placeholder` text appears when the value is empty.

## Progress Bars

There's no built-in progress element yet, but a helper function does the job:

```go
func progressBar(value, width int) string {
    filled := value * width / 100
    bar := ""
    for i := 0; i < width; i++ {
        if i < filled {
            bar += "█"
        } else {
            bar += "░"
        }
    }
    return bar
}
```

Then use it in your template with styling:

```gsx
<div class="flex gap-2 items-center">
    <span class="font-dim w-10">Download:</span>
    <span class="text-cyan">{progressBar(e.progress.Get(), 25)}</span>
    <span class="text-cyan font-bold">{fmt.Sprintf("%d%%", e.progress.Get())}</span>
</div>
```

Color the bar with `text-cyan`, `text-green`, `text-yellow`, etc. to convey meaning (progress, success, warning).

## Complete Example

This elements gallery demonstrates every built-in element type in a scrollable layout:

```gsx
package main

import (
    "fmt"
    tui "github.com/grindlemire/go-tui"
)

type elementsApp struct {
    progress *tui.State[int]
    scrollY  *tui.State[int]
    content  *tui.Ref
}

func Elements() *elementsApp {
    return &elementsApp{
        progress: tui.NewState(62),
        scrollY:  tui.NewState(0),
        content:  tui.NewRef(),
    }
}

func (e *elementsApp) scrollBy(delta int) {
    el := e.content.El()
    if el == nil {
        return
    }
    _, maxY := el.MaxScroll()
    newY := e.scrollY.Get() + delta
    if newY < 0 {
        newY = 0
    } else if newY > maxY {
        newY = maxY
    }
    e.scrollY.Set(newY)
}

func (e *elementsApp) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
        tui.OnRune('q', func(ke tui.KeyEvent) { ke.App().Stop() }),
        tui.OnRune('+', func(ke tui.KeyEvent) {
            v := e.progress.Get() + 5
            if v > 100 {
                v = 100
            }
            e.progress.Set(v)
        }),
        tui.OnRune('-', func(ke tui.KeyEvent) {
            v := e.progress.Get() - 5
            if v < 0 {
                v = 0
            }
            e.progress.Set(v)
        }),
        tui.OnRune('j', func(ke tui.KeyEvent) { e.scrollBy(1) }),
        tui.OnRune('k', func(ke tui.KeyEvent) { e.scrollBy(-1) }),
        tui.OnKey(tui.KeyDown, func(ke tui.KeyEvent) { e.scrollBy(1) }),
        tui.OnKey(tui.KeyUp, func(ke tui.KeyEvent) { e.scrollBy(-1) }),
    }
}

func (e *elementsApp) HandleMouse(me tui.MouseEvent) bool {
    switch me.Button {
    case tui.MouseWheelUp:
        e.scrollBy(-1)
        return true
    case tui.MouseWheelDown:
        e.scrollBy(1)
        return true
    }
    return false
}

func progressBar(value, width int) string {
    filled := value * width / 100
    bar := ""
    for i := 0; i < width; i++ {
        if i < filled {
            bar += "█"
        } else {
            bar += "░"
        }
    }
    return bar
}

templ (e *elementsApp) Render() {
    <div
        ref={e.content}
        class="flex-col gap-1 p-2 h-full"
        scrollable={tui.ScrollVertical}
        scrollOffset={0, e.scrollY.Get()}
    >
        <span class="text-gradient-cyan-magenta font-bold">Built-in Elements</span>

        // Text Elements
        <div class="flex-col border-rounded p-1 gap-1">
            <span class="text-gradient-cyan-magenta font-bold">Text Elements</span>
            <p>{"Paragraph text (<p>) wraps automatically when the content exceeds the available width. This demonstrates how longer text content is displayed."}</p>
            <hr />
            <span class="text-cyan">{"This is a <span> element for inline styled text"}</span>
            <br />
            <span class="font-dim">{"<hr> above draws a line, <br> inserts a blank line"}</span>
        </div>

        // Lists and Table side by side
        <div class="flex gap-1">
            <div class="flex-col border-rounded p-1 gap-1">
                <span class="text-gradient-cyan-magenta font-bold">{"Lists (<ul> / <li>)"}</span>
                <ul class="flex-col p-1">
                    <li><span>First item</span></li>
                    <li><span>Second item</span></li>
                    <li><span>Third item</span></li>
                    <li><span class="text-cyan">Fourth (styled)</span></li>
                </ul>
            </div>
            <div class="flex-col border-rounded p-1 gap-1">
                <span class="text-gradient-cyan-magenta font-bold">Table</span>
                <table class="flex-col p-1">
                    <div class="flex gap-2">
                        <span class="w-10 font-bold">Name</span>
                        <span class="w-10 font-bold">Role</span>
                        <span class="w-5 font-bold">Lvl</span>
                    </div>
                    <hr />
                    <div class="flex gap-2">
                        <span class="w-10 text-cyan">Alice</span>
                        <span class="w-10">Engineer</span>
                        <span class="w-5 text-green">Sr</span>
                    </div>
                    <div class="flex gap-2">
                        <span class="w-10 text-cyan">Bob</span>
                        <span class="w-10">Designer</span>
                        <span class="w-5 text-yellow">Jr</span>
                    </div>
                    <div class="flex gap-2">
                        <span class="w-10 text-cyan">Carol</span>
                        <span class="w-10">Manager</span>
                        <span class="w-5 text-green">Sr</span>
                    </div>
                </table>
            </div>
        </div>

        // Buttons
        <div class="flex-col border-rounded p-1 gap-1">
            <span class="text-gradient-cyan-magenta font-bold">Buttons</span>
            <div class="flex gap-2">
                <button>{"Save"}</button>
                <button>{"Cancel"}</button>
                <button class="font-bold">{"Submit"}</button>
                <button disabled={true}>{"Disabled"}</button>
            </div>
        </div>

        // Progress bars
        <div class="flex-col border-rounded p-1 gap-1">
            <span class="text-gradient-cyan-magenta font-bold">Progress Bars</span>
            <div class="flex gap-2 items-center">
                <span class="font-dim w-10">Download:</span>
                <span class="text-cyan">{progressBar(e.progress.Get(), 25)}</span>
                <span class="text-cyan font-bold">{fmt.Sprintf("%d%%", e.progress.Get())}</span>
            </div>
            <div class="flex gap-2 items-center">
                <span class="font-dim w-10">Upload:</span>
                <span class="text-green">{progressBar(100, 25)}</span>
                <span class="text-green font-bold">{"100%"}</span>
            </div>
            <div class="flex gap-2 items-center">
                <span class="font-dim w-10">Build:</span>
                <span class="text-yellow">{progressBar(35, 25)}</span>
                <span class="text-yellow font-bold">{"35%"}</span>
            </div>
        </div>

        <span class="font-dim">+/- adjust progress | j/k scroll | q quit</span>
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
        tui.WithRootComponent(Elements()),
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

Scroll through the gallery with j/k or arrow keys:

![Built-in Elements screenshot](/guides/05.png)

## Next Steps

- [Streaming Data](streaming) - Build a live data viewer with channels and auto-scroll
- [Refs and Click Handling](refs-and-clicks) - Mouse hit-testing with element references
