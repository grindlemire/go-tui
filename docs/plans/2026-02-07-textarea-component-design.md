# TextArea Component Design

## Overview

Promote the example-specific `TextArea` from `examples/ai-chat/textarea.go` to a framework-level component in the root `tui` package.

## Design Decisions

| Decision | Choice |
|----------|--------|
| Border | Optional (default none), configurable via `WithTextAreaBorder()` |
| Submit key | Configurable: `KeyEnter` (default) or `KeyCtrlEnter` |
| OnChange callback | No — YAGNI, clients can access `Text()` directly |
| Focusable interface | Yes — enables FocusManager integration and Tab navigation |
| Placeholder | Yes — with configurable style (defaults to dim) |

## Public API

```go
// TextArea is a multi-line text input with word wrapping and cursor management.
type TextArea struct { /* private fields */ }

// Constructor
func NewTextArea(opts ...TextAreaOption) *TextArea

// TextAreaOption configures a TextArea.
type TextAreaOption func(*TextArea)

// Sizing
func WithTextAreaWidth(cells int) TextAreaOption
func WithTextAreaMaxHeight(rows int) TextAreaOption

// Visual
func WithTextAreaBorder(b BorderStyle) TextAreaOption
func WithTextAreaTextStyle(s Style) TextAreaOption
func WithTextAreaPlaceholder(text string) TextAreaOption
func WithTextAreaPlaceholderStyle(s Style) TextAreaOption  // defaults to dim
func WithTextAreaCursor(r rune) TextAreaOption             // defaults to '▌'

// Behavior
func WithTextAreaSubmitKey(k Key) TextAreaOption           // KeyEnter (default) or KeyCtrlEnter
func WithTextAreaOnSubmit(fn func(string)) TextAreaOption

// State access
func (t *TextArea) Text() string
func (t *TextArea) SetText(s string)
func (t *TextArea) Clear()

// Interface implementations
var _ Component = (*TextArea)(nil)        // Render() *Element
var _ KeyListener = (*TextArea)(nil)      // KeyMap() KeyMap
var _ WatcherProvider = (*TextArea)(nil)  // Watchers() []Watcher
var _ Focusable = (*TextArea)(nil)        // IsFocusable, Focus, Blur, HandleEvent
```

## Internal Structure

```go
type TextArea struct {
    // Configuration (set via options, immutable after construction)
    width            int
    maxHeight        int
    border           BorderStyle
    textStyle        Style
    placeholder      string
    placeholderStyle Style
    cursorRune       rune
    submitKey        Key
    onSubmit         func(string)

    // Reactive state
    text      *State[string]
    cursorPos *State[int]
    blink     *State[bool]
    focused   *State[bool]
}
```

## Key Handling

- `OnRunesStop` for character input
- Configurable submit key (Enter vs Ctrl+Enter)
- The other key inserts newline
- Navigation: arrows, Home, End
- Editing: Backspace, Delete

## File Organization

New files in root `tui` package:
- `textarea.go` — TextArea struct, NewTextArea, methods
- `textarea_options.go` — TextAreaOption type and With* functions

## Client Usage

```gsx
type chat struct {
    textarea *tui.TextArea
}

func NewChat() *chat {
    c := &chat{}
    c.textarea = tui.NewTextArea(
        tui.WithTextAreaWidth(60),
        tui.WithTextAreaBorder(tui.BorderRounded),
        tui.WithTextAreaPlaceholder("Type a message..."),
        tui.WithTextAreaOnSubmit(c.handleSubmit),
    )
    return c
}

templ (c *chat) Render() {
    <div class="flex-col">
        @c.textarea
    </div>
}

func (c *chat) KeyMap() tui.KeyMap       { return c.textarea.KeyMap() }
func (c *chat) Watchers() []tui.Watcher  { return c.textarea.Watchers() }
```

## Migration

1. Create `tui/textarea.go` and `tui/textarea_options.go`
2. Update `examples/ai-chat/chat.gsx` to use `tui.TextArea`
3. Delete `examples/ai-chat/textarea.go`
4. Verify `@component` syntax works in code generator
