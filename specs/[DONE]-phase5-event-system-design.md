# Phase 5: Event System Specification

**Status:** Planned
**Version:** 1.0
**Last Updated:** 2025-01-24

---

## 1. Overview

### Purpose

The Event System provides the foundation for interactive TUI applications: reading keyboard/resize events from the terminal and dispatching them to focusable elements. It enables applications to respond to user input while maintaining smooth animations through a polling-based event loop.

Currently, the go-tui library provides:
- Terminal rendering (Phase 1)
- Flexbox layout engine (Phase 2)
- Element API with visual properties (Phase 3)
- Jitter-free layout rounding (Phase 4)

But keyboard input requires manual goroutine management (as seen in the dashboard example). The Event System will provide:
- Structured event types with rich key information
- Polling-based event reading with configurable timeouts
- A `Focusable` interface for elements that receive input
- A `FocusManager` for managing focus state across the element tree

### Goals

- **Polling-based event loop**: `PollEvent(timeout)` returns events or times out, enabling single-threaded animation + input handling
- **Structured event types**: `KeyEvent`, `ResizeEvent` with full modifier and special key support
- **Focusable interface**: Clean separation between layout elements and interactive elements
- **Manual focus navigation**: User controls when focus moves via `Next()`/`Prev()`/`SetFocus()`
- **No allocation on hot path**: Event reading should not allocate per-event
- **Testable design**: Mock event sources for unit testing interactive components

### Non-Goals

- Automatic Tab/Shift+Tab navigation (user handles navigation logic)
- Mouse support (keyboard-first; mouse deferred to future phase)
- Widget state management (Focusable handles its own state)
- Built-in widgets like List, Input (future phase builds on this foundation)
- Async event channels (polling provides sufficient flexibility)

---

## 2. Architecture

### Directory Structure

```
pkg/tui/
├── event.go            # Event interface and concrete event types
├── event_test.go
├── key.go              # Key and Modifier constants
├── key_test.go
├── reader.go           # EventReader interface and implementation
├── reader_test.go
├── focus.go            # Focusable interface and FocusManager
├── focus_test.go
├── app.go              # App type with integrated event loop
├── app_test.go
└── ... (existing files)
```

### Component Overview

| Component | Purpose |
|-----------|---------|
| `event.go` | Event interface and types: `KeyEvent`, `ResizeEvent` |
| `key.go` | Key constants (arrows, enter, escape) and Modifier flags |
| `reader.go` | `EventReader` for reading/polling events from terminal |
| `focus.go` | `Focusable` interface and `FocusManager` for focus state |
| `app.go` | `App` type that owns terminal, buffer, focus, and event reader |

### Data Flow

```
┌─────────────────────────────────────────────────────────────────┐
│  User Code (Application Loop)                                   │
│  for {                                                          │
│      event, ok := app.PollEvent(50 * time.Millisecond)          │
│      if ok { app.Dispatch(event) }                              │
│      // animate, update state                                   │
│      app.Render()                                               │
│  }                                                              │
└─────────────────────────────┬───────────────────────────────────┘
                              │ PollEvent()
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│  EventReader                                                    │
│  - Reads from terminal stdin                                    │
│  - Parses ANSI escape sequences into KeyEvent                   │
│  - Detects terminal resize → ResizeEvent                        │
│  - Returns (Event, bool) with timeout support                   │
└─────────────────────────────┬───────────────────────────────────┘
                              │ Dispatch()
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│  FocusManager                                                   │
│  - Tracks currently focused Focusable                           │
│  - Dispatches events to focused element                         │
│  - User calls Next()/Prev()/SetFocus() to change focus          │
└─────────────────────────────┬───────────────────────────────────┘
                              │ HandleEvent()
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│  Focusable Element                                              │
│  - Receives event via HandleEvent()                             │
│  - Returns true if handled, false to propagate                  │
│  - Updates internal state                                       │
└─────────────────────────────────────────────────────────────────┘
```

### Integration with Element API

The Event System integrates with the existing Element API through the `Focusable` interface:

```
                    ┌─────────────────┐
                    │ layout.Layoutable │
                    └────────┬────────┘
                             │
                    ┌────────┴────────┐
                    │ element.Element │
                    └────────┬────────┘
                             │ embeds
              ┌──────────────┼──────────────┐
              │              │              │
     ┌────────┴────────┐ ┌───┴───┐ ┌────────┴────────┐
     │  element.Text   │ │  ...  │ │ FocusableElement │
     └─────────────────┘ └───────┘ └────────┬────────┘
                                            │ implements
                                   ┌────────┴────────┐
                                   │ tui.Focusable   │
                                   └─────────────────┘
```

User-defined focusable elements embed `*Element` and implement `Focusable`.

---

## 3. Core Entities

### 3.1 Event Interface

```go
// pkg/tui/event.go

// Event is the base interface for all terminal events.
// Use type switch to handle specific event types.
type Event interface {
    // isEvent is a marker method to prevent external implementations.
    isEvent()
}
```

### 3.2 KeyEvent

```go
// KeyEvent represents a keyboard input event.
type KeyEvent struct {
    // Key is the key pressed. For printable characters, this is KeyRune.
    // For special keys (arrows, function keys), this is the specific constant.
    Key Key

    // Rune is the character for KeyRune events. Zero for special keys.
    Rune rune

    // Mod contains modifier flags (Ctrl, Alt, Shift).
    Mod Modifier
}

func (KeyEvent) isEvent() {}

// Common query methods for ergonomic usage

// IsRune returns true if this is a printable character event.
func (e KeyEvent) IsRune() bool { return e.Key == KeyRune }

// Is checks if the event matches a specific key with optional modifiers.
// Example: event.Is(KeyEnter) or event.Is(KeyRune, ModCtrl)
func (e KeyEvent) Is(key Key, mods ...Modifier) bool

// Char returns the rune if this is a KeyRune event, or 0 otherwise.
func (e KeyEvent) Char() rune
```

### 3.3 Key Constants

```go
// pkg/tui/key.go

// Key represents a keyboard key.
type Key uint16

const (
    // KeyNone represents no key (zero value).
    KeyNone Key = iota

    // KeyRune represents a printable character. Check KeyEvent.Rune for the character.
    KeyRune

    // Special keys
    KeyEscape
    KeyEnter
    KeyTab
    KeyBackspace
    KeyDelete
    KeyInsert

    // Arrow keys
    KeyUp
    KeyDown
    KeyLeft
    KeyRight

    // Navigation keys
    KeyHome
    KeyEnd
    KeyPageUp
    KeyPageDown

    // Function keys
    KeyF1
    KeyF2
    KeyF3
    KeyF4
    KeyF5
    KeyF6
    KeyF7
    KeyF8
    KeyF9
    KeyF10
    KeyF11
    KeyF12

    // Special combinations that map to control codes
    KeyCtrlSpace // Ctrl+Space = NUL (0x00)
    KeyCtrlA     // through KeyCtrlZ for Ctrl+letter
    KeyCtrlZ
)

// Modifier represents keyboard modifier flags.
type Modifier uint8

const (
    ModNone  Modifier = 0
    ModCtrl  Modifier = 1 << iota
    ModAlt
    ModShift
)

// Has checks if the modifier set includes the given modifier.
func (m Modifier) Has(mod Modifier) bool { return m&mod != 0 }

// String returns a human-readable representation.
func (m Modifier) String() string
```

**Design Rationale**:
- **Separate Key and Rune**: Distinguishes special keys from printable characters cleanly
- **Modifier as bitfield**: Efficient storage and combination checking
- **Explicit Ctrl+letter keys**: `KeyCtrlA` through `KeyCtrlZ` for common shortcuts, while `ModCtrl + KeyRune` handles arbitrary Ctrl+char combinations

### 3.4 ResizeEvent

```go
// ResizeEvent is emitted when the terminal is resized.
type ResizeEvent struct {
    Width  int
    Height int
}

func (ResizeEvent) isEvent() {}
```

### 3.5 EventReader

```go
// pkg/tui/reader.go

// EventReader reads events from the terminal.
// It is designed for polling-based event loops.
type EventReader interface {
    // PollEvent reads the next event with a timeout.
    // Returns (event, true) if an event was read, or (nil, false) on timeout.
    // A timeout of 0 performs a non-blocking check.
    // A negative timeout blocks indefinitely.
    PollEvent(timeout time.Duration) (Event, bool)

    // Close releases resources. Must be called when done.
    Close() error
}

// stdinReader implements EventReader for a real terminal.
type stdinReader struct {
    fd      int           // stdin file descriptor
    buf     []byte        // Read buffer for escape sequences
    pending []Event       // Parsed events waiting to be returned
    sigCh   chan os.Signal // For SIGWINCH (resize) handling
}

// NewEventReader creates an EventReader for the given terminal.
// The terminal should already be in raw mode.
func NewEventReader(in *os.File) (EventReader, error)
```

**Implementation Notes**:

1. **Non-blocking reads**: Use `syscall.Select` or `poll` with timeout to avoid blocking indefinitely
2. **Escape sequence parsing**: Buffer partial sequences; timeout incomplete sequences as individual bytes
3. **Resize detection**: Listen for `SIGWINCH` signal on Unix systems
4. **Buffer reuse**: Pre-allocate read buffer to avoid per-read allocations

### 3.6 Escape Sequence Parsing

```go
// Internal to stdinReader

// parseInput parses buffered bytes into events.
// Handles:
// - Single printable characters → KeyEvent{Key: KeyRune, Rune: r}
// - Control characters (0x00-0x1F) → KeyEvent{Key: KeyCtrlA + offset}
// - CSI sequences (\x1b[...) → Arrow keys, function keys with modifiers
// - SS3 sequences (\x1bO...) → Some function keys
func (r *stdinReader) parseInput() []Event

// Escape sequence patterns:
// \x1b[A     → KeyUp
// \x1b[B     → KeyDown
// \x1b[C     → KeyRight
// \x1b[D     → KeyLeft
// \x1b[1;2A  → KeyUp + ModShift (2 = shift)
// \x1b[1;3A  → KeyUp + ModAlt (3 = alt)
// \x1b[1;5A  → KeyUp + ModCtrl (5 = ctrl)
// \x1b[H     → KeyHome
// \x1b[F     → KeyEnd
// \x1b[5~    → KeyPageUp
// \x1b[6~    → KeyPageDown
// \x1bOP     → KeyF1 (SS3)
// \x1b[15~   → KeyF5 (CSI)
```

**Modifier Encoding** (xterm standard):
```
Modifier byte = 1 + (shift ? 1 : 0) + (alt ? 2 : 0) + (ctrl ? 4 : 0)

1 = none
2 = shift
3 = alt
4 = shift + alt
5 = ctrl
6 = ctrl + shift
7 = ctrl + alt
8 = ctrl + alt + shift
```

### 3.7 Focusable Interface

```go
// pkg/tui/focus.go

// Focusable is implemented by elements that can receive keyboard focus.
// Elements implementing Focusable should embed *element.Element and
// add input handling logic.
type Focusable interface {
    // IsFocusable returns whether this element can currently receive focus.
    // May return false for disabled elements.
    IsFocusable() bool

    // HandleEvent processes a keyboard event.
    // Returns true if the event was consumed, false to allow propagation.
    HandleEvent(event Event) bool

    // Focus is called when this element gains focus.
    // Implementations typically update visual state (e.g., highlight border).
    Focus()

    // Blur is called when this element loses focus.
    // Implementations typically revert visual state.
    Blur()
}
```

**Design Rationale**:
- **Separate interface**: Not all Elements need focus capability; clean separation
- **Boolean return**: Allows event bubbling when element doesn't handle an event
- **Focus/Blur hooks**: Enable visual feedback without coupling to render logic

### 3.8 FocusManager

```go
// FocusManager tracks focus state for a set of focusable elements.
// It does NOT automatically handle Tab navigation; the user controls
// when focus moves by calling Next(), Prev(), or SetFocus().
type FocusManager struct {
    elements []Focusable  // Registered focusable elements in order
    current  int          // Index of currently focused element (-1 = none)
}

// NewFocusManager creates a FocusManager with the given focusable elements.
// The first focusable element is focused by default.
func NewFocusManager(elements ...Focusable) *FocusManager

// Register adds a focusable element to the manager.
func (f *FocusManager) Register(elem Focusable)

// Unregister removes a focusable element from the manager.
func (f *FocusManager) Unregister(elem Focusable)

// Focused returns the currently focused element, or nil if none.
func (f *FocusManager) Focused() Focusable

// SetFocus moves focus to the specified element.
// Does nothing if the element is not registered or not focusable.
func (f *FocusManager) SetFocus(elem Focusable)

// Next moves focus to the next focusable element.
// Wraps around to the first element if at the end.
// Does nothing if there are no focusable elements.
func (f *FocusManager) Next()

// Prev moves focus to the previous focusable element.
// Wraps around to the last element if at the beginning.
func (f *FocusManager) Prev()

// Dispatch sends an event to the currently focused element.
// Returns true if the event was handled.
func (f *FocusManager) Dispatch(event Event) bool
```

**Design Rationale**:
- **Manual navigation**: User decides when to call `Next()`/`Prev()`, giving full control over focus flow
- **Linear order**: Elements are focused in registration order; for complex layouts, register in desired tab order
- **Dispatch convenience**: Centralized event routing to focused element

---

## 4. App Type

The `App` type brings together terminal, buffer, event reader, focus manager, and root element into a cohesive application object.

```go
// pkg/tui/app.go

// App manages the application lifecycle: terminal setup, event loop, and rendering.
type App struct {
    terminal *ANSITerminal
    buffer   *Buffer
    reader   EventReader
    focus    *FocusManager
    root     *element.Element  // nil until SetRoot is called
}

// NewApp creates a new application with the given terminal.
// The terminal is put into raw mode and alternate screen mode.
func NewApp() (*App, error)

// Close restores the terminal to its original state.
// Must be called when the application exits.
func (a *App) Close() error

// SetRoot sets the root element tree for rendering.
func (a *App) SetRoot(root *element.Element)

// Root returns the current root element.
func (a *App) Root() *element.Element

// Size returns the current terminal size.
func (a *App) Size() (width, height int)

// Focus returns the FocusManager for this app.
func (a *App) Focus() *FocusManager

// PollEvent reads the next event with a timeout.
// Convenience wrapper around the EventReader.
func (a *App) PollEvent(timeout time.Duration) (Event, bool)

// Dispatch sends an event to the focused element.
// Handles ResizeEvent internally by updating buffer size.
// Returns true if the event was consumed.
func (a *App) Dispatch(event Event) bool

// Render clears the buffer, renders the element tree, and flushes to terminal.
func (a *App) Render()
```

### 4.1 Typical Application Loop

```go
func main() {
    app, err := tui.NewApp()
    if err != nil {
        log.Fatal(err)
    }
    defer app.Close()

    // Build element tree
    root := element.New(/* ... */)
    app.SetRoot(root)

    // Register focusable elements
    input1 := NewMyInput("Name")
    input2 := NewMyInput("Email")
    app.Focus().Register(input1)
    app.Focus().Register(input2)

    // Main loop
    for {
        event, ok := app.PollEvent(50 * time.Millisecond)
        if ok {
            switch e := event.(type) {
            case tui.KeyEvent:
                if e.Key == tui.KeyEscape {
                    return // Exit on Escape
                }
                if e.Key == tui.KeyTab {
                    app.Focus().Next()
                } else {
                    app.Dispatch(event)
                }
            }
        }

        // Update state, animate, etc.
        // ...

        app.Render()
    }
}
```

---

## 5. User Experience

### 5.1 Creating a Focusable Element

```go
// Example: A simple focusable input element

type Input struct {
    *element.Element
    value    string
    cursor   int
    focused  bool
}

func NewInput(placeholder string) *Input {
    i := &Input{
        Element: element.New(
            element.WithWidth(30),
            element.WithHeight(1),
            element.WithBorder(tui.BorderSingle),
        ),
    }
    return i
}

// Implement tui.Focusable

func (i *Input) IsFocusable() bool { return true }

func (i *Input) HandleEvent(event tui.Event) bool {
    ke, ok := event.(tui.KeyEvent)
    if !ok {
        return false
    }

    switch ke.Key {
    case tui.KeyRune:
        i.value = i.value[:i.cursor] + string(ke.Rune) + i.value[i.cursor:]
        i.cursor++
        return true
    case tui.KeyBackspace:
        if i.cursor > 0 {
            i.value = i.value[:i.cursor-1] + i.value[i.cursor:]
            i.cursor--
        }
        return true
    case tui.KeyLeft:
        if i.cursor > 0 {
            i.cursor--
        }
        return true
    case tui.KeyRight:
        if i.cursor < len(i.value) {
            i.cursor++
        }
        return true
    }
    return false
}

func (i *Input) Focus() {
    i.focused = true
    i.SetBorderStyle(tui.NewStyle().Foreground(tui.Cyan))
}

func (i *Input) Blur() {
    i.focused = false
    i.SetBorderStyle(tui.NewStyle().Foreground(tui.White))
}
```

### 5.2 Event Handling Patterns

```go
// Pattern 1: Simple key check
if ke.Key == tui.KeyEnter {
    submit()
}

// Pattern 2: Check with modifier
if ke.Is(tui.KeyRune, tui.ModCtrl) && ke.Char() == 's' {
    save()
}

// Pattern 3: Type switch for multiple event types
switch e := event.(type) {
case tui.KeyEvent:
    handleKey(e)
case tui.ResizeEvent:
    resize(e.Width, e.Height)
}

// Pattern 4: Event dispatch with fallback
if !app.Dispatch(event) {
    // Event not handled by focused element
    handleGlobalKey(event)
}
```

### 5.3 Before/After Comparison

**Before (current dashboard example):**
```go
// Manual goroutine for keypress
done := make(chan struct{})
go func() {
    b := make([]byte, 1)
    os.Stdin.Read(b)
    close(done)
}()

ticker := time.NewTicker(50 * time.Millisecond)
for {
    select {
    case <-done:
        return
    case <-ticker.C:
        // render...
    }
}
```

**After (with Event System):**
```go
for {
    event, ok := app.PollEvent(50 * time.Millisecond)
    if ok {
        if ke, isKey := event.(tui.KeyEvent); isKey && ke.Key == tui.KeyEscape {
            return
        }
        app.Dispatch(event)
    }
    // render...
    app.Render()
}
```

---

## 6. Implementation Details

### 6.1 Non-Blocking Read on Unix

```go
// Use select() syscall with timeout for non-blocking reads
func (r *stdinReader) PollEvent(timeout time.Duration) (Event, bool) {
    // Return pending events first
    if len(r.pending) > 0 {
        ev := r.pending[0]
        r.pending = r.pending[1:]
        return ev, true
    }

    // Check for resize signal
    select {
    case <-r.sigCh:
        w, h := getTerminalSize()
        return ResizeEvent{Width: w, Height: h}, true
    default:
    }

    // Set up select() with timeout
    var tv syscall.Timeval
    if timeout >= 0 {
        tv.Sec = int64(timeout / time.Second)
        tv.Usec = int64((timeout % time.Second) / time.Microsecond)
    }

    var readFds syscall.FdSet
    readFds.Set(r.fd)

    n, err := syscall.Select(r.fd+1, &readFds, nil, nil, &tv)
    if err != nil || n == 0 {
        return nil, false
    }

    // Read available bytes
    nRead, err := syscall.Read(r.fd, r.buf)
    if err != nil || nRead == 0 {
        return nil, false
    }

    // Parse into events
    r.pending = r.parseInput(r.buf[:nRead])
    if len(r.pending) > 0 {
        ev := r.pending[0]
        r.pending = r.pending[1:]
        return ev, true
    }

    return nil, false
}
```

### 6.2 Escape Sequence State Machine

```go
// Parser states for escape sequence decoding
type parseState int

const (
    stateGround parseState = iota
    stateEscape           // Got ESC
    stateCSI              // Got ESC [
    stateCSIParam         // Reading CSI parameters
    stateSS3              // Got ESC O
)

func (r *stdinReader) parseInput(data []byte) []Event {
    var events []Event
    state := stateGround
    var params []int
    var currentParam int

    for _, b := range data {
        switch state {
        case stateGround:
            if b == 0x1b {
                state = stateEscape
            } else if b < 0x20 {
                // Control character
                events = append(events, KeyEvent{Key: controlToKey(b)})
            } else {
                // Printable character (UTF-8 handling needed)
                events = append(events, KeyEvent{Key: KeyRune, Rune: rune(b)})
            }

        case stateEscape:
            if b == '[' {
                state = stateCSI
                params = params[:0]
                currentParam = 0
            } else if b == 'O' {
                state = stateSS3
            } else {
                // Alt+key
                events = append(events, KeyEvent{Key: KeyRune, Rune: rune(b), Mod: ModAlt})
                state = stateGround
            }

        case stateCSI:
            // Parse CSI sequence...
            // ...
        }
    }

    return events
}
```

### 6.3 UTF-8 Handling

```go
// Multi-byte UTF-8 runes require buffering across reads
func (r *stdinReader) decodeUTF8(data []byte) ([]rune, []byte) {
    var runes []rune
    for len(data) > 0 {
        r, size := utf8.DecodeRune(data)
        if r == utf8.RuneError && size == 1 {
            // Incomplete sequence, keep for next read
            break
        }
        runes = append(runes, r)
        data = data[size:]
    }
    return runes, data // Return remaining incomplete bytes
}
```

---

## 7. Testing Strategy

### 7.1 MockEventReader

```go
// MockEventReader is an EventReader for testing.
type MockEventReader struct {
    events []Event
    index  int
}

func NewMockEventReader(events ...Event) *MockEventReader {
    return &MockEventReader{events: events}
}

func (m *MockEventReader) PollEvent(timeout time.Duration) (Event, bool) {
    if m.index >= len(m.events) {
        return nil, false
    }
    ev := m.events[m.index]
    m.index++
    return ev, true
}

func (m *MockEventReader) Close() error { return nil }
```

### 7.2 Test Categories

| Category | Test Focus |
|----------|------------|
| `key_test.go` | Key constants, modifier operations, string representations |
| `event_test.go` | Event type assertions, KeyEvent helper methods |
| `reader_test.go` | Escape sequence parsing, UTF-8 decoding, timeout behavior |
| `focus_test.go` | Focus state transitions, registration/unregistration, dispatch |
| `app_test.go` | App lifecycle, resize handling, render integration |

### 7.3 Escape Sequence Tests

```go
func TestParseArrowKeys(t *testing.T) {
    type tc struct {
        input    []byte
        expected []KeyEvent
    }
    tests := map[string]tc{
        "up":         {[]byte("\x1b[A"), []KeyEvent{{Key: KeyUp}}},
        "down":       {[]byte("\x1b[B"), []KeyEvent{{Key: KeyDown}}},
        "with_shift": {[]byte("\x1b[1;2A"), []KeyEvent{{Key: KeyUp, Mod: ModShift}}},
        "with_ctrl":  {[]byte("\x1b[1;5A"), []KeyEvent{{Key: KeyUp, Mod: ModCtrl}}},
    }

    for name, tt := range tests {
        t.Run(name, func(t *testing.T) {
            reader := &stdinReader{buf: make([]byte, 256)}
            events := reader.parseInput(tt.input)
            // Assert events match expected...
        })
    }
}
```

---

## 8. Performance Considerations

### 8.1 Allocation-Free Hot Path

| Operation | Target | Approach |
|-----------|--------|----------|
| PollEvent | Zero alloc | Reuse read buffer, pre-allocate pending slice |
| Parse escape | Zero alloc | Event structs are value types (no pointers) |
| Dispatch | Zero alloc | No intermediate collections |

### 8.2 Read Buffer Size

- Default buffer: 256 bytes (handles typical escape sequences)
- Maximum escape sequence: ~20 bytes (CSI with parameters)
- UTF-8 continuation: Buffer incomplete bytes between reads

### 8.3 Event Throughput

Target: Handle rapid key repeat (30+ keys/second) without dropped events.

- Use non-blocking select() to avoid blocking on read
- Buffer multiple parsed events in `pending` slice
- Process all buffered data before returning

---

## 9. Platform Considerations

### 9.1 Unix (Linux, macOS)

- **Raw mode**: Use `termios` via `golang.org/x/term` or direct syscall
- **Resize signal**: `SIGWINCH` via `signal.Notify()`
- **Non-blocking read**: `syscall.Select()` or `poll()`

### 9.2 Windows (Deferred)

Windows Console API differs significantly:
- Events via `ReadConsoleInput` (not byte stream)
- No ANSI escape sequences in older versions (requires VT mode)
- Resize via event, not signal

**Recommendation**: Defer Windows support to a future phase. Design the `EventReader` interface to enable platform-specific implementations.

---

## 10. Complexity Assessment

| Factor | Assessment |
|--------|------------|
| New files | 5 files in `pkg/tui/` |
| Core algorithm complexity | Moderate—escape sequence parsing is finicky but well-documented |
| Integration points | App integrates terminal, buffer, focus, and reader |
| Testing scope | Significant—many edge cases in escape sequences |
| Platform considerations | Unix-focused; Windows deferred |

**Assessed Size:** Medium
**Recommended Phases:** 3

**Rationale:** The Event System has moderate complexity with well-defined boundaries. Three phases allow for incremental delivery:
1. **Core types and parsing**: Event types, Key constants, escape sequence parser
2. **EventReader and FocusManager**: Polling implementation, focus state management
3. **App integration**: App type, resize handling, integration tests

---

## 11. Success Criteria

1. `PollEvent(timeout)` returns events within timeout, or `(nil, false)` on timeout
2. All common keys are correctly parsed: arrows, function keys, enter, escape, backspace
3. Modifiers (Ctrl, Alt, Shift) are correctly detected on supported terminals
4. UTF-8 characters are correctly decoded into `KeyRune` events
5. `ResizeEvent` is generated when terminal is resized
6. `FocusManager.Next()`/`Prev()`/`SetFocus()` correctly manage focus
7. `FocusManager.Dispatch()` routes events to the focused `Focusable`
8. Focus/Blur hooks are called on focus transitions
9. `App.Render()` integrates with Element tree rendering
10. Dashboard example can be rewritten using the new event loop pattern

---

## 12. Open Questions

1. **Should PollEvent use channels internally?**
   → No. Direct syscall-based polling is simpler and more predictable. Channels add overhead and goroutine management.

2. **How to handle terminals that don't support certain sequences?**
   → Best effort parsing. Unknown sequences are discarded or returned as raw bytes. Document known compatibility issues.

3. **Should FocusManager support nested focus (e.g., focus groups)?**
   → Deferred. Linear focus order is sufficient for v1. Nested focus can be added later.

4. **Should ResizeEvent be dispatched to focused element?**
   → No. Resize is handled at the App level (updates buffer size). Focused elements don't need resize events.

5. **How to handle clipboard (Ctrl+C/Ctrl+V)?**
   → Deferred. Clipboard support requires platform-specific APIs. For now, Ctrl+C generates interrupt signal (raw mode can capture it as `KeyCtrlC`).
