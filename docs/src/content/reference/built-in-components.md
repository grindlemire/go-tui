# Built-in Components Reference

## Overview

go-tui ships with three built-in components: `Input` (single-line text), `TextArea` (multi-line text), and `Modal` (overlay dialog). Input and TextArea handle text entry, cursor management, and focus. Modal handles backdrop rendering, focus trapping, and preemptive key blocking.

All three implement `Component`, `KeyListener`, and `AppBinder`. Use them standalone or embed them in a larger UI.

## Input

A single-line text input with cursor management, horizontal scrolling, and placeholder support.

```go
import tui "github.com/grindlemire/go-tui"

inp := tui.NewInput(
    tui.WithInputWidth(30),
    tui.WithInputBorder(tui.BorderRounded),
    tui.WithInputPlaceholder("Type here..."),
    tui.WithInputOnSubmit(func(text string) {
        // handle submitted text
    }),
)
```

### NewInput

```go
func NewInput(opts ...InputOption) *Input
```

Creates a new `Input` with the given options. Default values:

| Setting     | Default      | Description                              |
|-------------|--------------|------------------------------------------|
| Width       | 20           | Characters visible before scrolling      |
| Border      | `BorderNone` | No border                                |
| TextStyle   | `Style{}`    | Default terminal style                   |
| Placeholder | `""`         | No placeholder text                      |
| PlaceholderStyle | `Style{}.Dim()` | Dim text for placeholder          |
| Cursor      | `'▌'`        | Block cursor character                   |
| FocusColor  | `Cyan`       | Border color when focused                |

### Reactive Value Binding

Bind the Input to a `*State[string]` for two-way binding. The Input shares the state directly, so typing updates the state and changing the state updates the display:

```go
name := tui.NewState("")
inp := tui.NewInput(
    tui.WithInputValue(name),
    tui.WithInputBorder(tui.BorderRounded),
)
// Later: name.Set("Alice") updates the input display
// Typing in the input updates name.Get()
```

In `.gsx`:

```gsx
<input value={s.name} placeholder="Type your name..." border={tui.BorderRounded} />
```

### GSX Attributes

All `<input>` attributes and their types:

| Attribute | Type | Description |
|-----------|------|-------------|
| `value` | `*State[string]` | Two-way text binding |
| `placeholder` | `string` | Text shown when empty and unfocused |
| `placeholderStyle` | `tui.Style` | Placeholder styling (default: dim) |
| `width` | `int` | Width in characters (default 20) |
| `border` | `tui.BorderStyle` | Border style |
| `textStyle` | `tui.Style` | Text styling |
| `cursor` | `rune` | Cursor character (default '▌') |
| `focusColor` | `tui.Color` | Border color when focused (default Cyan) |
| `borderGradient` | `tui.Gradient` | Border gradient when unfocused |
| `focusGradient` | `tui.Gradient` | Border gradient when focused |
| `onSubmit` | `func(string)` | Called when Enter is pressed |
| `onChange` | `func(string)` | Called when text changes |
| `autoFocus` | `bool` | Focus this input on startup |

### Focus Border Styling

Control how the border looks when focused and unfocused:

```go
// Solid color when focused (default: Cyan)
tui.WithInputFocusColor(tui.Magenta)

// Gradient border when unfocused
tui.WithInputBorderGradient(tui.NewGradient(tui.Blue, tui.Cyan))

// Gradient border when focused (overrides focusColor)
tui.WithInputFocusGradient(tui.NewGradient(tui.Cyan, tui.Magenta))
```

In `.gsx`:

```gsx
<input
    value={s.query}
    border={tui.BorderRounded}
    focusColor={tui.Magenta}
    borderGradient={tui.NewGradient(tui.Blue, tui.Cyan)}
    focusGradient={tui.NewGradient(tui.Cyan, tui.Magenta)}
/>
```

### Keyboard Behavior

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
| Home        | Move cursor to start              |
| End         | Move cursor to end                |

**Submit:**

| Key         | Action                            |
|-------------|-----------------------------------|
| Enter       | Trigger `onSubmit` callback       |

### InputOption Functions

| Function | Description |
|----------|-------------|
| `WithInputWidth(int)` | Width in characters (default 20) |
| `WithInputBorder(BorderStyle)` | Border style |
| `WithInputTextStyle(Style)` | Text style |
| `WithInputPlaceholder(string)` | Placeholder text |
| `WithInputPlaceholderStyle(Style)` | Placeholder style (default: dim) |
| `WithInputCursor(rune)` | Cursor character (default '▌') |
| `WithInputValue(*State[string])` | Reactive two-way text binding |
| `WithInputFocusColor(Color)` | Border color when focused (default Cyan) |
| `WithInputBorderGradient(Gradient)` | Border gradient when unfocused |
| `WithInputFocusGradient(Gradient)` | Border gradient when focused |
| `WithInputOnSubmit(func(string))` | Enter key callback |
| `WithInputOnChange(func(string))` | Text change callback |

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
| FocusColor  | `Cyan`       | Border color when focused                |
| SubmitKey   | `KeyEnter`   | Enter submits, Ctrl+J inserts newline    |

### GSX Attributes

All `<textarea>` attributes and their types:

| Attribute | Type | Description |
|-----------|------|-------------|
| `value` | `*State[string]` | Two-way text binding |
| `placeholder` | `string` | Text shown when empty and unfocused |
| `placeholderStyle` | `tui.Style` | Placeholder styling (default: dim) |
| `width` | `int` | Width in characters (default 40) |
| `maxHeight` | `int` | Maximum visible rows (0 = unlimited) |
| `border` | `tui.BorderStyle` | Border style |
| `textStyle` | `tui.Style` | Text styling |
| `cursor` | `rune` | Cursor character (default '▌') |
| `focusColor` | `tui.Color` | Border color when focused (default Cyan) |
| `borderGradient` | `tui.Gradient` | Border gradient when unfocused |
| `focusGradient` | `tui.Gradient` | Border gradient when focused |
| `submitKey` | `tui.Key` | Key that triggers submit (default KeyEnter) |
| `onSubmit` | `func(string)` | Called when submit key is pressed |
| `autoFocus` | `bool` | Focus this text area on startup |

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

### TextAreaOption Functions

Options follow the functional options pattern. Each returns a `TextAreaOption` (which is `func(*TextArea)`).

#### WithTextAreaWidth

```go
func WithTextAreaWidth(cells int) TextAreaOption
```

Sets the width in characters. Text wraps at this boundary. Default: 40.

```go
ta := tui.NewTextArea(tui.WithTextAreaWidth(80))
```

#### WithTextAreaMaxHeight

```go
func WithTextAreaMaxHeight(rows int) TextAreaOption
```

Caps the visible height to `rows` lines of text. Set to 0 (the default) for unlimited height. This does not include border rows. If a border is set, the actual element height is `maxHeight + 2`.

```go
ta := tui.NewTextArea(tui.WithTextAreaMaxHeight(10))
```

#### WithTextAreaBorder

```go
func WithTextAreaBorder(b BorderStyle) TextAreaOption
```

Sets the border style around the text area. Default: `BorderNone`.

```go
ta := tui.NewTextArea(tui.WithTextAreaBorder(tui.BorderRounded))
```

#### WithTextAreaTextStyle

```go
func WithTextAreaTextStyle(s Style) TextAreaOption
```

Sets the style for the text content. Default: zero-value `Style{}` (terminal default).

```go
ta := tui.NewTextArea(
    tui.WithTextAreaTextStyle(tui.NewStyle().Foreground(tui.ANSIColor(tui.Cyan))),
)
```

#### WithTextAreaPlaceholder

```go
func WithTextAreaPlaceholder(text string) TextAreaOption
```

Sets placeholder text shown when the TextArea is empty and unfocused. Default: empty string (no placeholder).

```go
ta := tui.NewTextArea(tui.WithTextAreaPlaceholder("Enter your message..."))
```

#### WithTextAreaPlaceholderStyle

```go
func WithTextAreaPlaceholderStyle(s Style) TextAreaOption
```

Sets the style for placeholder text. Default: `Style{}.Dim()`.

```go
ta := tui.NewTextArea(
    tui.WithTextAreaPlaceholderStyle(tui.NewStyle().Dim().Italic()),
)
```

#### WithTextAreaCursor

```go
func WithTextAreaCursor(r rune) TextAreaOption
```

Sets the cursor character. Default: `'▌'` (left half block).

```go
ta := tui.NewTextArea(tui.WithTextAreaCursor('█'))
```

#### WithTextAreaSubmitKey

```go
func WithTextAreaSubmitKey(k Key) TextAreaOption
```

Sets which key triggers the `onSubmit` callback. Default: `KeyEnter`.

When the submit key is `KeyEnter`, pressing Enter triggers submit and Ctrl+J inserts a newline. For any other submit key, Enter inserts a newline and the configured key triggers submit.

```go
// Ctrl+S submits, Enter inserts newlines (good for multi-line editing)
ta := tui.NewTextArea(tui.WithTextAreaSubmitKey(tui.KeyCtrlS))
```

#### WithTextAreaOnSubmit

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

#### WithTextAreaValue

```go
func WithTextAreaValue(state *State[string]) TextAreaOption
```

Binds the TextArea to a `*State[string]` for two-way binding. The TextArea shares the state directly, so parent components can read or change the text at any time.

```go
note := tui.NewState("")
ta := tui.NewTextArea(tui.WithTextAreaValue(note))
// note.Get() reflects whatever the user types
// note.Set("preset text") updates the textarea display
```

In `.gsx`:

```gsx
<textarea value={s.note} placeholder="Write a note..." border={tui.BorderRounded} />
```

#### WithTextAreaFocusColor

```go
func WithTextAreaFocusColor(c Color) TextAreaOption
```

Sets the border color when focused. Default: `Cyan`. Only visible when a border is set.

```go
ta := tui.NewTextArea(
    tui.WithTextAreaBorder(tui.BorderRounded),
    tui.WithTextAreaFocusColor(tui.Magenta),
)
```

##### WithTextAreaBorderGradient

```go
func WithTextAreaBorderGradient(g Gradient) TextAreaOption
```

Sets a gradient for the border color when unfocused. Only visible when a border is set.

```go
ta := tui.NewTextArea(
    tui.WithTextAreaBorder(tui.BorderRounded),
    tui.WithTextAreaBorderGradient(tui.NewGradient(tui.Blue, tui.Cyan)),
)
```

#### WithTextAreaFocusGradient

```go
func WithTextAreaFocusGradient(g Gradient) TextAreaOption
```

Sets a gradient for the border color when focused. Takes priority over `focusColor` when set.

```go
ta := tui.NewTextArea(
    tui.WithTextAreaBorder(tui.BorderRounded),
    tui.WithTextAreaFocusGradient(tui.NewGradient(tui.Cyan, tui.Magenta)),
)
```

### Example

A note-taking input using `TextArea` with a rounded border and Ctrl+S to submit:

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

## Modal

Modal renders a full-screen overlay with a backdrop effect, focus trapping, and preemptive key handling.

### Constructor

```go
func NewModal(opts ...ModalOption) *Modal
```

### ModalOption Functions

| Function | Description |
|----------|-------------|
| `WithModalOpen(state *State[bool])` | Bind visibility to a boolean state (required) |
| `WithModalBackdrop(b string)` | Backdrop style: `"dim"` (default), `"blank"`, or `"none"` |
| `WithModalCloseOnEscape(v bool)` | Escape closes the modal (default `true`) |
| `WithModalCloseOnBackdropClick(v bool)` | Backdrop click closes the modal (default `true`) |
| `WithModalTrapFocus(v bool)` | Restrict Tab navigation to modal children (default `true`) |
| `WithModalElementOptions(opts ...Option)` | Pass layout options to the overlay container (used by generated code for `class` attributes) |

### GSX Attributes

All `<modal>` attributes and their types:

| Attribute | Type | Description |
|-----------|------|-------------|
| `open` | `*State[bool]` | Controls visibility (required) |
| `backdrop` | `string` | `"dim"` (default), `"blank"`, or `"none"` |
| `closeOnEscape` | `bool` | Escape closes the modal (default true) |
| `closeOnBackdropClick` | `bool` | Backdrop click closes the modal (default true) |
| `trapFocus` | `bool` | Tab/Shift+Tab restricted to modal children (default true) |
| `class` | `string` | Tailwind classes for positioning (e.g. `"justify-center items-center"`) |

### Behavior

When open, the modal:

- Applies the backdrop effect (dim, blank, or none) to the buffer before rendering the overlay
- Traps Tab/Shift+Tab within its focusable children
- Handles Enter by calling `Activate()` on the focused element
- Blocks all parent key handlers via preemptive dispatch
- Closes on Escape (if `closeOnEscape` is true)
- Closes on backdrop click (if `closeOnBackdropClick` is true)
- Walks clicked elements up to find `onActivate` callbacks for mouse support

When closed, it returns a hidden placeholder element with no key bindings.

### Interfaces Implemented

| Interface | Purpose |
|-----------|---------|
| `Component` | `Render(app *App) *Element` returns the overlay element |
| `KeyListener` | `KeyMap()` returns Escape, Tab, Enter, and catch-all bindings |
| `MouseListener` | `HandleMouse()` handles backdrop click and onActivate delegation |
| `AppBinder` | `BindApp()` wires the open state to the app |

### GSX Usage

```gsx
<modal open={s.showDialog} class="justify-center items-center" backdrop="dim">
    <div class="border-rounded p-2 flex-col gap-1 w-40">
        <span class="font-bold">Title</span>
        <button class="px-2 border-rounded focusable" onActivate={s.onConfirm}>OK</button>
    </div>
</modal>
```

The `class` attribute on `<modal>` controls how the dialog is positioned within the full-screen overlay. Use `justify-center items-center` for a centered dialog or `justify-end items-stretch` for a bottom sheet.

### Inline Mode

Modals are not supported in inline mode (`WithInlineHeight`). The overlay system requires a full-screen buffer for backdrop effects, centering, and mouse hit testing. Modal overlays registered while in inline mode are silently ignored.

To show a modal from an inline app, switch to the alternate screen first:

```go
app.EnterAlternateScreen()
s.showDialog.Set(true)
// ... user interacts with modal ...
// on close:
app.ExitAlternateScreen()
```

See the [Inline Mode Guide](../guides/15-inline-mode) for the full pattern.

## Cross-References

- [Component Interfaces Reference](interfaces.md) — `Component`, `KeyListener`, `WatcherProvider`, `Focusable`, `AppBinder`
- [Events Reference](events.md) — `KeyEvent`, `Key` constants, `KeyMap`
- [State Reference](state.md) — `State[T]` used internally by TextArea
- [Styling Reference](styling.md) — `Style` and `BorderStyle` for visual configuration
- [Focus Guide](../guides/13-focus.md) — Focus management and Tab navigation
- [Inline Mode Guide](../guides/12-inline-mode.md) — Using TextArea in inline mode
