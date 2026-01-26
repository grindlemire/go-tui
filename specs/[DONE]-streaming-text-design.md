# Streaming Text Specification

**Status:** Planned\
**Version:** 1.0\
**Last Updated:** 2025-01-25

---

## 1. Overview

### Purpose

Enable streaming text content from external processes into scrollable TUI elements. This allows real-time display of process output, log streams, or any asynchronous text source within the go-tui framework.

### Goals

- Provide a channel-based interface (`chan string`) for streaming text into elements
- Implement smart auto-scroll behavior that follows new content unless user scrolls up
- Handle multi-line strings (split on newlines and render each line separately)
- Create a self-contained example demonstrating the streaming capability
- Integrate seamlessly with existing scrollable element infrastructure

### Non-Goals

- ANSI escape code parsing (plain text only for v1)
- io.Reader or callback-based interfaces (channel only for v1)
- Horizontal scrolling for long lines (vertical scrolling only)
- Text search or filtering within streamed content

---

## 2. Architecture

### Directory Structure

```
pkg/tui/element/
├── element.go         # MODIFY: Add onUpdate hook
├── options.go         # MODIFY: Add WithOnUpdate option
└── scroll.go          # Existing scroll implementation

examples/
└── streaming/
    └── main.go        # NEW: Streaming demo (defines StreamBox locally)
```

### Component Overview

| Component | Purpose |
|-----------|---------|
| `pkg/tui/element/element.go` | Add `onUpdate` hook called before each render |
| `pkg/tui/element/options.go` | Add `WithOnUpdate` option |
| `examples/streaming/main.go` | Demo with local StreamBox showing streaming into scrollable viewport |

### Flow Diagram

```
┌─────────────────┐     chan string     ┌─────────────────┐
│ External Source │────────────────────►│   StreamBox     │
│ (goroutine)     │                     │                 │
└─────────────────┘                     │ - Receives text │
                                        │ - Splits lines  │
                                        │ - Auto-scrolls  │
                                        └────────┬────────┘
                                                 │
                                                 ▼
                                        ┌─────────────────┐
                                        │    Element      │
                                        │ (scrollable)    │
                                        │                 │
                                        │ - Line children │
                                        │ - Scroll state  │
                                        └────────┬────────┘
                                                 │
                                                 ▼
                                        ┌─────────────────┐
                                        │  TUI Render     │
                                        └─────────────────┘
```

---

## 3. Core Entities

### Element onUpdate Hook

A new pre-render hook on Element that enables custom update logic:

```go
// In Element struct (element.go):
type Element struct {
    // ... existing fields ...

    // Pre-render hook for custom update logic (polling, animations, etc.)
    onUpdate func()
}

// SetOnUpdate sets a function called before each render.
// Useful for polling channels, updating animations, etc.
func (e *Element) SetOnUpdate(fn func()) {
    e.onUpdate = fn
}

// In Element.Render() - call hook before layout:
func (e *Element) Render(buf *tui.Buffer, width, height int) {
    if e.onUpdate != nil {
        e.onUpdate()
    }
    // ... rest of existing render logic
}
```

### WithOnUpdate Option

```go
// In options.go:

// WithOnUpdate sets a function called before each render.
func WithOnUpdate(fn func()) Option {
    return func(e *Element) {
        e.onUpdate = fn
    }
}
```

### Example StreamBox Pattern

The example demonstrates a `StreamBox` pattern that users can adapt:

```go
// StreamBox wraps an Element to provide channel-based text streaming.
// Defined locally in the example - not part of the library.
type StreamBox struct {
    elem        *element.Element
    textCh      <-chan string
    textStyle   tui.Style
    autoScroll  bool
}

func NewStreamBox(textCh <-chan string) *StreamBox {
    s := &StreamBox{
        elem:       element.New(
            element.WithScrollable(element.ScrollVertical),
            element.WithDirection(layout.Column),
        ),
        textCh:     textCh,
        autoScroll: true,
    }

    // Set up automatic polling via onUpdate hook
    s.elem.SetOnUpdate(s.poll)

    return s
}

func (s *StreamBox) Element() *element.Element {
    return s.elem
}

// poll is called automatically before each render
func (s *StreamBox) poll() {
    for {
        select {
        case text, ok := <-s.textCh:
            if !ok {
                return  // Channel closed
            }
            s.appendText(text)
        default:
            return  // No more messages available
        }
    }
}
```

This approach:
- **Library stays minimal** - only adds the onUpdate hook
- **Example shows the pattern** - users can copy/adapt as needed
- Polling happens automatically during render
- No background goroutines needed

### Auto-Scroll Behavior

The auto-scroll mechanism tracks whether the user has scrolled up:

1. **Initial state**: `autoScroll = true`
2. **On new content**: If `autoScroll == true`, scroll to bottom
3. **On user scroll up**: Set `autoScroll = false` (stop following)
4. **On user scroll to bottom**: Set `autoScroll = true` (resume following)

Detection logic:
- User is "at bottom" when: `scrollY >= maxScrollY` (within 1 line tolerance)
- User scrolled up when: `scrollY < maxScrollY - 1` after receiving scroll event

---

## 4. User Experience

### Library API (onUpdate hook)

```go
// Create an element with custom update logic
elem := element.New(
    element.WithScrollable(element.ScrollVertical),
    element.WithOnUpdate(func() {
        // Called before each render - poll channels, update state, etc.
    }),
)
```

### Example Usage (StreamBox pattern)

```go
// Create a channel for streaming
textCh := make(chan string)

// Create StreamBox (defined locally in example)
streamBox := NewStreamBox(textCh)
streamBox.elem.SetBorder(tui.BorderSingle)

// Add to element tree
parent.AddChild(streamBox.Element())

// Stream content from goroutine
go func() {
    defer close(textCh)
    for _, line := range lines {
        textCh <- line
        time.Sleep(100 * time.Millisecond)
    }
}()

// Main loop is clean - no polling code needed!
for {
    event, ok := app.PollEvent(50 * time.Millisecond)
    if ok {
        // ... handle events
    }
    app.Render()  // StreamBox polling happens automatically via onUpdate
}
```

### Example Output

```
┌────────────────────────────────────────┐
│ Streaming Demo - Press ESC to exit     │
├────────────────────────────────────────┤
│ [12:00:01] Process started...          │
│ [12:00:02] Initializing components...  │
│ [12:00:03] Loading configuration...    │
│ [12:00:04] Connecting to service...    │
│ [12:00:05] Ready to accept requests    │
│ [12:00:06] Processing request #1       │
│ [12:00:07] Processing request #2       │▒
│ [12:00:08] Processing request #3       │▒
│ [12:00:09] Processing request #4       │█ <-- User sees latest
└────────────────────────────────────────┘
```

---

## 5. Complexity Assessment

| Size | Phases | When to Use |
|------|--------|-------------|
| Small | 1-2 | Single component, bug fix, minor enhancement |
| Medium | 3-4 | New feature touching multiple files/components |
| Large | 5-6 | Cross-cutting feature, new subsystem |

**Assessed Size:** Small
**Recommended Phases:** 2
**Rationale:**
- Small modification to Element: add `onUpdate` hook field + `SetOnUpdate()` method
- Small modification to render: call hook before layout
- Add `WithOnUpdate` option
- One example file demonstrating the StreamBox pattern
- Library stays minimal - example shows how to compose streaming behavior

---

## 6. Success Criteria

1. Element's `onUpdate` hook is called before each render (enables extensibility)
2. `WithOnUpdate` option allows setting the hook at construction time
3. Example defines a local `StreamBox` type that uses the hook for automatic polling
4. Example receives text via channel and displays each line as a child element
5. Multi-line strings are properly split and displayed as separate lines
6. Auto-scroll follows new content when user is at the bottom
7. Auto-scroll pauses when user scrolls up, resumes when they scroll to bottom
8. Example demonstrates a simulated streaming process with visible auto-scroll behavior
9. User can manually scroll through streamed content using existing scroll keys

---

## 7. Open Questions

1. Should the example include max line management?
   → **Optional** - keep example simple, users can add if needed

2. How to handle partial lines (text without trailing newline)?
   → Example can buffer until newline received, or just add as-is
