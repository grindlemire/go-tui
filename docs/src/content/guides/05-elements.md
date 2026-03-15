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

`<input>` is a single-line text input with cursor management and placeholder support. It is self-closing and focus-aware.

Bind `value` to a `*State[string]` for two-way binding. Typing updates the state, and changing the state updates the display:

```gsx
<input value={s.name} placeholder="Type your name..." width={30} border={tui.BorderRounded} />
```

You can also set `focusColor` to change the border color when focused, or use `focusGradient` and `borderGradient` for gradient borders:

```gsx
<input
    value={s.name}
    placeholder="Type here..."
    width={30}
    border={tui.BorderRounded}
    focusColor={tui.Magenta}
/>

// Gradient border that shifts color when focused
<input
    value={s.query}
    placeholder="Search..."
    border={tui.BorderRounded}
    borderGradient={tui.NewGradient(tui.Blue, tui.Cyan)}
    focusGradient={tui.NewGradient(tui.Cyan, tui.Magenta)}
/>
```

Wire up `onSubmit` for Enter key handling and `onChange` to react to each keystroke.

## TextArea Element

`<textarea>` is a multi-line text input with word wrapping and cursor navigation. Like `<input>`, it is self-closing and focus-aware.

Bind `value` to a `*State[string]` for two-way binding. Set `maxHeight` to cap the visible rows, and use `submitKey` to control how Enter behaves:

```gsx
<textarea
    value={s.note}
    placeholder="Write a note..."
    width={40}
    maxHeight={6}
    border={tui.BorderRounded}
    onSubmit={s.onSave}
    focusColor={tui.BrightRed}
/>
```

By default, Enter triggers `onSubmit` and Ctrl+J inserts a newline. Set `submitKey` to a different key (like `tui.KeyCtrlS`) to make Enter insert newlines instead.

The same border styling options from Input apply here: `focusColor`, `borderGradient`, and `focusGradient`.

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

## Modal Dialogs

The `<modal>` element renders as a full-screen overlay. When open, it dims the background, traps keyboard focus, and blocks parent key handlers. Escape and backdrop clicks close it by default.

Bind the `open` attribute to a `*State[bool]` to control visibility. Use `onActivate` on buttons to handle Enter key activation:

```gsx
<modal open={s.showDialog} class="justify-center items-center">
    <div class="border-rounded p-2 flex-col gap-1 w-40">
        <span class="font-bold text-yellow">Are you sure?</span>
        <button class="px-2 border-rounded focusable text-green font-bold" onActivate={s.cancel}>Cancel</button>
        <button class="px-2 border-rounded focusable text-red font-bold" onActivate={s.confirm}>Confirm</button>
    </div>
</modal>
```

The modal container uses flexbox layout, so `justify-center items-center` centers the dialog, while `justify-end items-stretch` pins it as a bottom sheet. Tab and Shift+Tab cycle between focusable children. Enter triggers the focused element's `onActivate` callback. Mouse clicks on elements with `onActivate` also trigger it.

Key attributes:

| Attribute | Default | Description |
|-----------|---------|-------------|
| `open` | (required) | `*State[bool]` controlling visibility |
| `backdrop` | `"dim"` | `"dim"`, `"blank"`, or `"none"` |
| `closeOnEscape` | `true` | Escape key closes the modal |
| `closeOnBackdropClick` | `true` | Clicking outside the dialog closes it |
| `trapFocus` | `true` | Restrict Tab navigation and block parent key handlers |

Focusable elements with borders get an automatic cyan highlight when focused. The first focusable child receives focus when the modal opens.

## Complete Example

This elements gallery demonstrates every built-in element type in a scrollable layout, including Input and TextArea with two-way value binding:

```gsx
package main

import (
    "fmt"
    tui "github.com/grindlemire/go-tui"
)

type elementsApp struct {
    progress    *tui.State[int]
    scrollY     *tui.State[int]
    content     *tui.Ref
    name        *tui.State[string]
    note        *tui.State[string]
    selectedBtn *tui.State[string]
    btnRefs     *tui.RefMap[string]
}

func Elements() *elementsApp {
    return &elementsApp{
        progress:    tui.NewState(62),
        scrollY:     tui.NewState(0),
        content:     tui.NewRef(),
        name:        tui.NewState(""),
        note:        tui.NewState(""),
        selectedBtn: tui.NewState(""),
        btnRefs:     tui.NewRefMap[string](),
    }
}

func (e *elementsApp) onNoteSubmit(text string) {
    e.note.Set(text)
}

func greeting(name string) string {
    if name == "" {
        return "Hello, World!"
    }
    return fmt.Sprintf("Hello, %s!", name)
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
        tui.On(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
        tui.On(tui.Rune('q'), func(ke tui.KeyEvent) { ke.App().Stop() }),
        tui.On(tui.KeyTab, func(ke tui.KeyEvent) { ke.App().FocusNext() }),
        tui.On(tui.KeyTab.Shift(), func(ke tui.KeyEvent) { ke.App().FocusPrev() }),
        tui.On(tui.Rune('+'), func(ke tui.KeyEvent) {
            v := e.progress.Get() + 5
            if v > 100 {
                v = 100
            }
            e.progress.Set(v)
        }),
        tui.On(tui.Rune('-'), func(ke tui.KeyEvent) {
            v := e.progress.Get() - 5
            if v < 0 {
                v = 0
            }
            e.progress.Set(v)
        }),
        tui.On(tui.Rune('j'), func(ke tui.KeyEvent) { e.scrollBy(1) }),
        tui.On(tui.Rune('k'), func(ke tui.KeyEvent) { e.scrollBy(-1) }),
        tui.On(tui.KeyDown, func(ke tui.KeyEvent) { e.scrollBy(1) }),
        tui.On(tui.KeyUp, func(ke tui.KeyEvent) { e.scrollBy(-1) }),
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

    if me.Button == tui.MouseLeft && me.Action == tui.MousePress {
        for name, el := range e.btnRefs.All() {
            if el != nil && el.ContainsPoint(me.X, me.Y) {
                if name != "Disabled" {
                    e.selectedBtn.Set(name)
                }
                return true
            }
        }
    }

    return false
}

var buttonLabels = []string{"Save", "Cancel", "Submit", "Disabled"}

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
        class="flex-col gap-1 h-full"
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
                <table class="p-1">
                    <tr>
                        <th>Name</th>
                        <th>Role</th>
                        <th>Lvl</th>
                    </tr>
                    <hr />
                    <tr>
                        <td class="text-cyan">Alice</td>
                        <td>Engineer</td>
                        <td class="text-green">Sr</td>
                    </tr>
                    <tr>
                        <td class="text-cyan">Bob</td>
                        <td>Designer</td>
                        <td class="text-yellow">Jr</td>
                    </tr>
                </table>
            </div>
        </div>

        // Buttons
        <div class="flex-col border-rounded p-1 gap-1">
            <span class="text-gradient-cyan-magenta font-bold">Buttons</span>
            <div class="flex gap-2">
                for _, label := range buttonLabels {
                    if label == "Disabled" {
                        <button ref={e.btnRefs} key={label} class="font-dim" disabled={true}>{label}</button>
                    } else {
                        if label == e.selectedBtn.Get() {
                            <button ref={e.btnRefs} key={label} class="font-bold text-cyan">{label}</button>
                        } else {
                            <button ref={e.btnRefs} key={label}>{label}</button>
                        }
                    }
                }
            </div>
        </div>

        // Input & TextArea
        <div class="flex-col border-rounded p-1 gap-1">
            <span class="text-gradient-cyan-magenta font-bold">Input & TextArea</span>
            <div class="flex gap-2">
                <div class="flex-col gap-1 w-1/2">
                    <div class="flex gap-2 items-center">
                        <span class="font-dim">Name:</span>
                        <input
                            placeholder="Type your name..."
                            value={e.name}
                            width={30}
                            border={tui.BorderRounded}
                            focusGradient={tui.NewGradient(tui.Cyan, tui.Magenta)}
                        />
                    </div>
                    <span class="text-cyan font-bold" width={30}>
                        {greeting(e.name.Get())}
                    </span>
                </div>
                <div class="flex-col gap-1 w-1/2">
                    <div class="flex gap-2 items-center">
                        <span class="font-dim">Note:</span>
                        <textarea
                            placeholder="Write a note..."
                            width={30}
                            maxHeight={4}
                            border={tui.BorderRounded}
                            onSubmit={e.onNoteSubmit}
                            focusColor={tui.BrightRed}
                        />
                    </div>
                    if e.note.Get() != "" {
                        <span class="text-cyan font-bold">{fmt.Sprintf("Saved: %s", e.note.Get())}</span>
                    }
                </div>
            </div>
            <span class="font-dim">Tab to cycle focus | Esc to blur | Enter submits note</span>
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

        <span class="font-dim">tab focus input | +/- progress | j/k scroll | q quit</span>
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
