# Inline Mode and Alternate Screen

## Overview

By default, go-tui takes over the full terminal using the alternate screen buffer. Your app gets a clean canvas, and when it exits, the user's previous terminal content reappears untouched. This works well for most apps, but sometimes you want your UI to coexist with normal terminal output. A chat input that sits at the bottom while messages scroll above, or a progress bar that doesn't erase your command history.

Inline mode gives you that. Instead of using the alternate screen, your app occupies a fixed number of rows at the bottom of the terminal. Everything above those rows behaves like a normal terminal: text scrolls up into scrollback, and users can scroll back through history with their terminal's native scroll.

You can also switch between the two modes at runtime. A chat app might run inline for typing, then jump to the alternate screen for a settings panel.

## Inline Mode

Enable inline mode by passing `tui.WithInlineHeight(rows)` when creating your app:

```go
package main

import (
    "fmt"
    "os"

    tui "github.com/grindlemire/go-tui"
)

func main() {
    app, err := tui.NewApp(
        tui.WithInlineHeight(3),
    )
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    defer app.Close()

    app.SetRootComponent(MyApp())

    if err := app.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

The number you pass is how many rows your widget occupies at the bottom of the terminal. The rest of the terminal remains available for scrollback content.

A few things change when inline mode is active:

- The alternate screen is **not** used, so terminal history is preserved
- Mouse events are **disabled** by default, which allows the user's native terminal scrollback to work
- If you need mouse events, add `tui.WithMouse()` explicitly

## PrintAbove

Inline mode becomes most useful when you print content above the widget. As your app runs, you push lines upward into the terminal's scrollback:

```gsx
package main

import (
    "fmt"
    tui "github.com/grindlemire/go-tui"
)

type chatInput struct {
    app      *tui.App
    textarea *tui.TextArea
}

func ChatInput() *chatInput {
    c := &chatInput{}
    c.textarea = tui.NewTextArea(
        tui.WithTextAreaWidth(60),
        tui.WithTextAreaBorder(tui.BorderRounded),
        tui.WithTextAreaPlaceholder("Type a message..."),
        tui.WithTextAreaOnSubmit(c.send),
    )
    return c
}

func (c *chatInput) BindApp(app *tui.App) {
    c.app = app
    c.textarea.BindApp(app)
}

func (c *chatInput) send(text string) {
    if text == "" {
        return
    }
    c.textarea.Clear()
    c.app.PrintAboveln("You: %s", text)
}

func (c *chatInput) KeyMap() tui.KeyMap {
    km := c.textarea.KeyMap()
    km = append(km,
        tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
    )
    return km
}

func (c *chatInput) Watchers() []tui.Watcher {
    return c.textarea.Watchers()
}

templ (c *chatInput) Render() {
    @c.textarea
}
```

Four methods handle printing above the widget:

| Method | Newline | Thread-safe |
|---|---|---|
| `PrintAbove(format, args...)` | No | No. Call from the event loop. |
| `PrintAboveln(format, args...)` | Yes | No. Call from the event loop. |
| `QueuePrintAbove(format, args...)` | No | Yes. Safe from any goroutine. |
| `QueuePrintAboveln(format, args...)` | Yes | Yes. Safe from any goroutine. |

`PrintAboveln` and `PrintAbove` work like `fmt.Sprintf`: pass a format string and arguments. The "ln" variants append a newline automatically.

Use the `Queue` variants when printing from a watcher callback, a goroutine, or anywhere outside the main event loop. They queue the write so it executes on the UI thread:

```go
// From a goroutine or watcher
go func() {
    result := fetchData()
    app.QueuePrintAboveln("Received: %s", result)
}()
```

For inserting fully rendered elements (tables, styled cards, templ component output) into the scrollback, use `PrintAboveElement`. See the [Inline Streaming Guide](/guides/inline-streaming#inserting-elements-mid-stream) for details.

## Dynamic Height

You can change the inline widget's height at runtime with `SetInlineHeight`. This is useful for text areas that grow as the user types:

```go
func (c *chatInput) updateHeight() {
    h := c.textarea.Height()
    if h < 3 {
        h = 3
    }
    c.app.SetInlineHeight(h)
}
```

Call `SetInlineHeight` from render methods or the event loop. The change takes effect immediately: the buffer resizes and the widget redraws at its new position.

`InlineHeight()` returns the current height (or 0 if the app is not in inline mode):

```go
current := app.InlineHeight()
```

The height is capped to the terminal height automatically, so you don't need to worry about requesting more rows than the terminal has.

## Alternate Screen

Even when running in inline mode, you can temporarily jump to the alternate screen for an overlay UI. Think of it as a full-screen modal:

```go
// Switch to full-screen
app.EnterAlternateScreen()

// ... user interacts with full-screen UI ...

// Return to inline mode
app.ExitAlternateScreen()
```

Three methods control this:

- `EnterAlternateScreen()` — saves the current inline state and switches to full-screen. The buffer resizes to fill the terminal. If already in alternate screen mode, this is a no-op.
- `ExitAlternateScreen()` — restores the saved inline state. The terminal's scrollback reappears, and the widget returns to its previous position. If not in alternate screen mode, this is a no-op.
- `IsInAlternateScreen()` — returns `true` if currently in alternate screen mode.

When you enter the alternate screen, the framework saves your inline height, start row, and layout state. When you exit, everything is restored and the widget gets a full redraw.

## Inline Startup Modes

When your inline app starts, existing content may be visible in the terminal. The startup mode controls what happens to that content:

```go
app, err := tui.NewApp(
    tui.WithInlineHeight(5),
    tui.WithInlineStartupMode(tui.InlineStartupFreshViewport),
)
```

Three modes are available:

| Mode | Behavior |
|---|---|
| `InlineStartupPreserveVisible` | Keeps existing visible rows. Previous content stays where it is. This is the **default**. |
| `InlineStartupFreshViewport` | Clears the visible viewport. Existing visible rows are discarded (not pushed to scrollback). |
| `InlineStartupSoftReset` | Pushes visible rows into scrollback via newline flow, then starts with a clean viewport. Previous content is preserved in scrollback. |

For most apps, the default (`PreserveVisible`) works fine. Use `FreshViewport` if you want a clean start without any previous terminal output visible. Use `SoftReset` if you want a clean start but still want the user to be able to scroll back and see what was on screen before your app launched.

## Combining Inline and Alternate Screen

A common pattern is running your main UI inline while using the alternate screen for overlays like settings panels or help screens. The ai-chat example in the repository demonstrates this pattern.

The key idea: track whether the overlay is showing with a `State[bool]`, and toggle between modes with a key binding:

```gsx
package main

import (
    tui "github.com/grindlemire/go-tui"
)

type myApp struct {
    app          *tui.App
    showSettings *tui.State[bool]
    textarea     *tui.TextArea
}

func MyApp() *myApp {
    a := &myApp{
        showSettings: tui.NewState(false),
    }
    a.textarea = tui.NewTextArea(
        tui.WithTextAreaWidth(60),
        tui.WithTextAreaBorder(tui.BorderRounded),
        tui.WithTextAreaPlaceholder("Type here..."),
        tui.WithTextAreaOnSubmit(a.send),
    )
    return a
}

func (a *myApp) BindApp(app *tui.App) {
    a.app = app
    a.showSettings.BindApp(app)
    a.textarea.BindApp(app)
}

func (a *myApp) send(text string) {
    if text == "" {
        return
    }
    a.textarea.Clear()
    a.app.PrintAboveln("You: %s", text)
}

func (a *myApp) toggleSettings() {
    if a.showSettings.Get() {
        _ = a.app.ExitAlternateScreen()
        a.showSettings.Set(false)
        return
    }
    a.showSettings.Set(true)
    _ = a.app.EnterAlternateScreen()
}

func (a *myApp) KeyMap() tui.KeyMap {
    if a.showSettings.Get() {
        return tui.KeyMap{
            tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { a.toggleSettings() }),
            tui.OnKey(tui.KeyCtrlC, func(ke tui.KeyEvent) { ke.App().Stop() }),
        }
    }

    km := a.textarea.KeyMap()
    km = append(km,
        tui.OnKeyStop(tui.KeyCtrlS, func(ke tui.KeyEvent) { a.toggleSettings() }),
        tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
        tui.OnKey(tui.KeyCtrlC, func(ke tui.KeyEvent) { ke.App().Stop() }),
    )
    return km
}

func (a *myApp) Watchers() []tui.Watcher {
    return a.textarea.Watchers()
}

templ (a *myApp) Render() {
    if a.showSettings.Get() {
        <div class="flex-col h-full p-1 border-rounded border-cyan">
            <span class="font-bold text-cyan">Settings</span>
            <span class="font-dim">Press Escape to return</span>
        </div>
    } else {
        @a.textarea
    }
}
```

The flow works like this:

1. App starts in inline mode with `WithInlineHeight(3)`
2. User types messages; `PrintAboveln` pushes them into scrollback
3. Ctrl+S calls `toggleSettings`, which sets state and calls `EnterAlternateScreen()`
4. The render method sees `showSettings` is true and renders the settings panel full-screen
5. Escape calls `toggleSettings` again, which calls `ExitAlternateScreen()` and restores the inline widget
6. The terminal scrollback is preserved through the round trip

The conditional `KeyMap` is important here. When settings are showing, you return different bindings than when the text area is active. This keeps the two modes from interfering with each other.

Here's what the inline text area looks like at the bottom of the terminal:

![Inline Mode and Alternate Screen screenshot](/guides/15.png)

**Cross-references**: [Events Guide](events), [State Guide](state)
