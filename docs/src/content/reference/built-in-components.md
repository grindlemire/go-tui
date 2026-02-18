# Built-in Components Reference

## Overview

go-tui ships with one built-in component: `TextArea`, a multi-line text input with word wrapping, cursor management, and keyboard navigation. It handles the common case of text entry so you don't have to build input handling from scratch.

`TextArea` implements five component interfaces: `Component`, `KeyListener`, `WatcherProvider`, `Focusable`, and `AppBinder`. You can use it as a standalone root component or embed it inside a larger UI.

## TextArea

A multi-line text input with word wrapping and a blinking cursor.

```go
import tui "github.com/grindlemire/go-tui"

ta := tui.NewTextArea(
    tui.WithTextAreaWidth(60),
    tui.WithTextAreaBorder(tui.BorderRounded),
    tui.WithTextAreaPlaceholder("Type something..."),
    tui.WithTextAreaOnSubmit(func(text string) {
        // handle submitted text
    }),
)
```

### NewTextArea

```go
func NewTextArea(opts ...TextAreaOption) *TextArea
```

Creates a new `TextArea` with the given options. Default values:

| Setting     | Default      | Description                              |
|-------------|--------------|------------------------------------------|
| Width       | 40           | Characters per line before wrapping      |
| MaxHeight   | 0 (no limit) | Maximum rows of text visible             |
| Border      | `BorderNone` | No border                                |
| TextStyle   | `Style{}`    | Default terminal style                   |
| Placeholder | `""`         | No placeholder text                      |
| PlaceholderStyle | `Style{}.Dim()` | Dim text for placeholder          |
| Cursor      | `'▌'`        | Block cursor character                   |
| SubmitKey   | `KeyEnter`   | Enter submits, Ctrl+J inserts newline    |

### State Access Methods

#### Text

```go
func (t *TextArea) Text() string
```

Returns the current text content.

#### SetText

```go
func (t *TextArea) SetText(s string)
```

Replaces the text and moves the cursor to the end.

```go
ta.SetText("Hello, world!")
fmt.Println(ta.Text()) // "Hello, world!"
```

#### Clear

```go
func (t *TextArea) Clear()
```

Removes all text and resets the cursor to position 0.

#### Height

```go
func (t *TextArea) Height() int
```

Returns the total rendered height in rows, including border rows if a border is set. The height depends on the current text content and word wrapping. If `maxHeight` is set, the returned value is capped to that limit (plus border rows).

### Focus Methods

`TextArea` implements the `Focusable` interface. When focused, the cursor blinks and keystrokes are captured. When unfocused, placeholder text appears (if configured) and no keystrokes are processed.

#### IsFocusable

```go
func (t *TextArea) IsFocusable() bool
```

Always returns `true`.

#### Focus

```go
func (t *TextArea) Focus()
```

Activates the text area. The cursor becomes visible and starts blinking.

#### Blur

```go
func (t *TextArea) Blur()
```

Deactivates the text area. The cursor disappears and placeholder text shows if the input is empty.

### HandleEvent

```go
func (t *TextArea) HandleEvent(e Event) bool
```

Processes a keyboard event against the TextArea's key map. Returns `true` if the event was handled (and propagation should stop), `false` otherwise. This method is part of the `Focusable` interface and is called automatically by the focus manager when the TextArea has focus.

### BindApp

```go
func (t *TextArea) BindApp(app *App)
```

Binds the TextArea's internal reactive states to the given `App`, so that state changes trigger re-renders. Called automatically when the TextArea is used as a root component or mounted as a sub-component.

### Keyboard Behavior

`TextArea` returns a `KeyMap` with built-in bindings. All bindings use stop propagation so keystrokes don't bubble up to parent components.

**Text input:**

| Key         | Action                            |
|-------------|-----------------------------------|
| Any rune    | Insert character at cursor        |
| Backspace   | Delete character before cursor    |
| Delete      | Delete character at cursor        |

**Navigation:**

| Key         | Action                            |
|-------------|-----------------------------------|
| Left        | Move cursor left                  |
| Right       | Move cursor right                 |
| Up          | Move cursor up one line           |
| Down        | Move cursor down one line         |
| Home        | Move cursor to start of line      |
| End         | Move cursor to end of line        |

**Submit and newline (default submit key = Enter):**

| Key         | Action                            |
|-------------|-----------------------------------|
| Enter       | Trigger `onSubmit` callback       |
| Ctrl+J      | Insert newline character          |

When `submitKey` is set to something other than `KeyEnter`, the behavior flips: Enter inserts a newline and the configured key triggers submit.

### Cursor Blink

`TextArea` implements `WatcherProvider` and returns a timer watcher that toggles the cursor visibility every 500ms while focused. The cursor resets to visible on every keystroke so the user always sees where they're typing.

## TextAreaOption Functions

Options follow the functional options pattern. Each returns a `TextAreaOption` (which is `func(*TextArea)`).

### WithTextAreaWidth

```go
func WithTextAreaWidth(cells int) TextAreaOption
```

Sets the width in characters. Text wraps at this boundary. Default: 40.

```go
ta := tui.NewTextArea(tui.WithTextAreaWidth(80))
```

### WithTextAreaMaxHeight

```go
func WithTextAreaMaxHeight(rows int) TextAreaOption
```

Caps the visible height to `rows` lines of text. Set to 0 (the default) for unlimited height. This does not include border rows — if a border is set, the actual element height is `maxHeight + 2`.

```go
ta := tui.NewTextArea(tui.WithTextAreaMaxHeight(10))
```

### WithTextAreaBorder

```go
func WithTextAreaBorder(b BorderStyle) TextAreaOption
```

Sets the border style around the text area. Default: `BorderNone`.

```go
ta := tui.NewTextArea(tui.WithTextAreaBorder(tui.BorderRounded))
```

### WithTextAreaTextStyle

```go
func WithTextAreaTextStyle(s Style) TextAreaOption
```

Sets the style for the text content. Default: zero-value `Style{}` (terminal default).

```go
ta := tui.NewTextArea(
    tui.WithTextAreaTextStyle(tui.NewStyle().Foreground(tui.ANSIColor(tui.Cyan))),
)
```

### WithTextAreaPlaceholder

```go
func WithTextAreaPlaceholder(text string) TextAreaOption
```

Sets placeholder text shown when the TextArea is empty and unfocused. Default: empty string (no placeholder).

```go
ta := tui.NewTextArea(tui.WithTextAreaPlaceholder("Enter your message..."))
```

### WithTextAreaPlaceholderStyle

```go
func WithTextAreaPlaceholderStyle(s Style) TextAreaOption
```

Sets the style for placeholder text. Default: `Style{}.Dim()`.

```go
ta := tui.NewTextArea(
    tui.WithTextAreaPlaceholderStyle(tui.NewStyle().Dim().Italic()),
)
```

### WithTextAreaCursor

```go
func WithTextAreaCursor(r rune) TextAreaOption
```

Sets the cursor character. Default: `'▌'` (left half block).

```go
ta := tui.NewTextArea(tui.WithTextAreaCursor('█'))
```

### WithTextAreaSubmitKey

```go
func WithTextAreaSubmitKey(k Key) TextAreaOption
```

Sets which key triggers the `onSubmit` callback. Default: `KeyEnter`.

When the submit key is `KeyEnter`, pressing Enter triggers submit and Ctrl+J inserts a newline. For any other submit key, Enter inserts a newline and the configured key triggers submit.

```go
// Ctrl+S submits, Enter inserts newlines (good for multi-line editing)
ta := tui.NewTextArea(tui.WithTextAreaSubmitKey(tui.KeyCtrlS))
```

### WithTextAreaOnSubmit

```go
func WithTextAreaOnSubmit(fn func(string)) TextAreaOption
```

Sets the callback invoked when the submit key is pressed. The callback receives the current text content.

```go
ta := tui.NewTextArea(
    tui.WithTextAreaOnSubmit(func(text string) {
        fmt.Println("Submitted:", text)
    }),
)
```

## Complete Example

A simple note-taking input using `TextArea` with a rounded border and Ctrl+S to submit:

```go
package main

import (
    "fmt"
    "os"

    tui "github.com/grindlemire/go-tui"
)

func main() {
    ta := tui.NewTextArea(
        tui.WithTextAreaWidth(60),
        tui.WithTextAreaMaxHeight(10),
        tui.WithTextAreaBorder(tui.BorderRounded),
        tui.WithTextAreaPlaceholder("Write a note..."),
        tui.WithTextAreaSubmitKey(tui.KeyCtrlS),
        tui.WithTextAreaOnSubmit(func(text string) {
            fmt.Printf("Saved: %s\n", text)
        }),
    )

    app, err := tui.NewApp(
        tui.WithRootComponent(ta),
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

For an inline chat pattern with `PrintAbove`, see the [Inline Mode Guide](../guides/12-inline-mode.md).

## Cross-References

- [Component Interfaces Reference](interfaces.md) — `Component`, `KeyListener`, `WatcherProvider`, `Focusable`, `AppBinder`
- [Events Reference](events.md) — `KeyEvent`, `Key` constants, `KeyMap`
- [State Reference](state.md) — `State[T]` used internally by TextArea
- [Styling Reference](styling.md) — `Style` and `BorderStyle` for visual configuration
- [Focus Guide](../guides/13-focus.md) — Focus management and Tab navigation
- [Inline Mode Guide](../guides/12-inline-mode.md) — Using TextArea in inline mode
