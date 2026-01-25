# Phase 5: Event System Implementation Plan

Implementation phases for the Event System. Each phase builds on the previous and has clear acceptance criteria.

---

## Phase 1: Core Event Types and Key Parsing

**Reference:** [phase5-event-system-design.md §3](./phase5-event-system-design.md#3-core-entities)

**Review:** false

**Completed in commit:** phase-1-complete

- [x] Create `pkg/tui/key.go`
  - Define `Key` type as `uint16`
  - Define all key constants: `KeyNone`, `KeyRune`, `KeyEscape`, `KeyEnter`, `KeyTab`, `KeyBackspace`, `KeyDelete`, `KeyInsert`
  - Define arrow keys: `KeyUp`, `KeyDown`, `KeyLeft`, `KeyRight`
  - Define navigation keys: `KeyHome`, `KeyEnd`, `KeyPageUp`, `KeyPageDown`
  - Define function keys: `KeyF1` through `KeyF12`
  - Define control keys: `KeyCtrlA` through `KeyCtrlZ`
  - Define `Modifier` type as `uint8` with `ModNone`, `ModCtrl`, `ModAlt`, `ModShift`
  - Implement `Modifier.Has(mod Modifier) bool` method
  - Implement `Key.String()` and `Modifier.String()` for debugging
  - See [design §3.3](./phase5-event-system-design.md#33-key-constants)

- [x] Create `pkg/tui/event.go`
  - Define `Event` interface with private marker method `isEvent()`
  - Define `KeyEvent` struct with `Key`, `Rune`, `Mod` fields
  - Implement `KeyEvent.isEvent()` marker
  - Implement `KeyEvent.IsRune() bool` helper
  - Implement `KeyEvent.Is(key Key, mods ...Modifier) bool` for ergonomic checks
  - Implement `KeyEvent.Char() rune` helper
  - Define `ResizeEvent` struct with `Width`, `Height` fields
  - Implement `ResizeEvent.isEvent()` marker
  - See [design §3.1, §3.2, §3.4](./phase5-event-system-design.md#31-event-interface)

- [x] Create `pkg/tui/parse.go` (internal escape sequence parsing)
  - Define `parseState` constants: `stateGround`, `stateEscape`, `stateCSI`, `stateCSIParam`, `stateSS3`
  - Implement `controlToKey(b byte) Key` for control characters (0x00-0x1F)
  - Implement `parseCSI(params []int, final byte) (Key, Modifier)` for CSI sequences
  - Implement `parseSS3(b byte) Key` for SS3 function key sequences
  - Implement `decodeModifier(param int) Modifier` for xterm modifier encoding
  - Implement `parseInput(data []byte) []Event` as the main parser function
  - Handle basic printable characters → `KeyEvent{Key: KeyRune, Rune: r}`
  - Handle control characters → `KeyEvent{Key: KeyCtrlA + offset}`
  - Handle arrow keys: `\x1b[A/B/C/D` → `KeyUp/Down/Right/Left`
  - Handle arrows with modifiers: `\x1b[1;2A` → `KeyUp + ModShift`
  - Handle navigation: `\x1b[H`, `\x1b[F`, `\x1b[5~`, `\x1b[6~`
  - Handle function keys (CSI): `\x1b[15~` → `KeyF5`, etc.
  - Handle function keys (SS3): `\x1bOP` → `KeyF1`, etc.
  - Handle Alt+key: `\x1b` + printable → `KeyRune + ModAlt`
  - See [design §3.6](./phase5-event-system-design.md#36-escape-sequence-parsing)

- [x] Create `pkg/tui/key_test.go`
  - Test `Modifier.Has()` with various combinations
  - Test `Key.String()` for all key constants
  - Test `Modifier.String()` for all modifier combinations

- [x] Create `pkg/tui/event_test.go`
  - Test `KeyEvent.IsRune()` for rune and non-rune events
  - Test `KeyEvent.Is()` with and without modifiers
  - Test `KeyEvent.Char()` for rune and non-rune events
  - Test type assertions for `Event` interface

- [x] Create `pkg/tui/parse_test.go`
  - Test `parseInput` with printable characters (ASCII)
  - Test `parseInput` with control characters (Ctrl+A through Ctrl+Z)
  - Test `parseInput` with arrow keys (all 4 directions)
  - Test `parseInput` with arrow keys + modifiers (Shift, Alt, Ctrl)
  - Test `parseInput` with function keys F1-F12 (both CSI and SS3 forms)
  - Test `parseInput` with navigation keys (Home, End, PageUp, PageDown)
  - Test `parseInput` with Alt+letter combinations
  - Test `parseInput` with partial sequences (should buffer)
  - Test `parseInput` with multiple events in single input

**Tests:** Run `go test ./pkg/tui/... -run "TestKey|TestEvent|TestParse"` once at phase end

---

## Phase 2: EventReader and FocusManager

**Reference:** [phase5-event-system-design.md §3.5, §3.7, §3.8](./phase5-event-system-design.md#35-eventreader)

**Review:** false

**Completed in commit:** phase-2-complete

- [x] Create `pkg/tui/reader.go`
  - Define `EventReader` interface with `PollEvent(timeout time.Duration) (Event, bool)` and `Close() error`
  - Define `stdinReader` struct with `fd`, `buf`, `pending`, `partialBuf`, `sigCh` fields
  - Implement `NewEventReader(in *os.File) (EventReader, error)`
    - Store file descriptor
    - Allocate read buffer (256 bytes)
    - Set up SIGWINCH signal channel for resize events
  - Implement `stdinReader.PollEvent(timeout time.Duration) (Event, bool)`
    - Return pending events first (from previous parse)
    - Check resize signal channel (non-blocking)
    - Use `syscall.Select()` with timeout for non-blocking stdin check
    - Read available bytes into buffer
    - Call `parseInput()` to decode events
    - Store extra events in `pending` slice
    - Return first event or (nil, false) on timeout
  - Implement `stdinReader.Close() error` to clean up signal channel
  - Handle UTF-8 continuation: buffer incomplete multi-byte sequences
  - See [design §3.5, §6.1](./phase5-event-system-design.md#35-eventreader)

- [x] Create `pkg/tui/reader_unix.go` (build tag: `//go:build unix`)
  - Implement `getTerminalSize() (width, height int)` using `syscall.TIOCGWINSZ`
  - Implement `fdSet.Set(fd int)` helper for `syscall.FdSet`
  - Implement `selectWithTimeout(fd int, timeout time.Duration) (ready bool, err error)` wrapper

- [x] Create `pkg/tui/focus.go`
  - Define `Focusable` interface
    - `IsFocusable() bool`
    - `HandleEvent(event Event) bool`
    - `Focus()`
    - `Blur()`
  - Define `FocusManager` struct with `elements []Focusable`, `current int`
  - Implement `NewFocusManager(elements ...Focusable) *FocusManager`
    - Store elements
    - Set `current = 0` if elements exist, else `-1`
    - Call `Focus()` on first element if exists
  - Implement `FocusManager.Register(elem Focusable)`
    - Append to elements slice
    - If first element, set current and call Focus()
  - Implement `FocusManager.Unregister(elem Focusable)`
    - Find and remove from slice
    - Adjust current index if needed
    - Call Blur() if removed element was focused
  - Implement `FocusManager.Focused() Focusable`
    - Return `elements[current]` or nil if `current == -1`
  - Implement `FocusManager.SetFocus(elem Focusable)`
    - Find element index
    - Call Blur() on current, Focus() on new
    - Update current
  - Implement `FocusManager.Next()` with wraparound
    - Call Blur() on current
    - Increment current (wrap to 0 at end)
    - Skip non-focusable elements
    - Call Focus() on new current
  - Implement `FocusManager.Prev()` with wraparound
    - Same as Next() but decrement (wrap to end at 0)
  - Implement `FocusManager.Dispatch(event Event) bool`
    - If no focused element, return false
    - Call `focused.HandleEvent(event)` and return result
  - See [design §3.7, §3.8](./phase5-event-system-design.md#37-focusable-interface)

- [x] Create `pkg/tui/mock_reader.go`
  - Define `MockEventReader` struct with `events []Event`, `index int`
  - Implement `NewMockEventReader(events ...Event) *MockEventReader`
  - Implement `PollEvent(timeout time.Duration) (Event, bool)` returning queued events
  - Implement `Close() error` as no-op
  - See [design §7.1](./phase5-event-system-design.md#71-mockeventreader)

- [x] Create `pkg/tui/reader_test.go`
  - Test `MockEventReader` returns queued events in order
  - Test `MockEventReader` returns (nil, false) when exhausted
  - Test `stdinReader` with mock data (via pipe) - basic key parsing
  - Test `stdinReader` timeout behavior (returns false when no data)
  - Test resize event generation (if testable without real terminal)

- [x] Create `pkg/tui/focus_test.go`
  - Create mock `Focusable` for testing (tracks Focus/Blur calls)
  - Test `NewFocusManager` focuses first element
  - Test `FocusManager.Next()` cycles through elements
  - Test `FocusManager.Prev()` cycles backward
  - Test `FocusManager.SetFocus()` changes focus correctly
  - Test `FocusManager.Register()` adds element
  - Test `FocusManager.Unregister()` removes element and adjusts focus
  - Test `FocusManager.Dispatch()` routes to focused element
  - Test focus wrapping at boundaries
  - Test skipping non-focusable elements in Next/Prev

**Tests:** Run `go test ./pkg/tui/... -run "TestReader|TestFocus|TestMock"` once at phase end

---

## Phase 3: App Integration and Examples

**Reference:** [phase5-event-system-design.md §4](./phase5-event-system-design.md#4-app-type)

**Review:** false

**Completed in commit:** phase-3-complete

- [x] Create `pkg/tui/app.go`
  - Define `App` struct with `terminal *ANSITerminal`, `buffer *Buffer`, `reader EventReader`, `focus *FocusManager`, `root Renderable`
  - Implement `NewApp() (*App, error)`
    - Create ANSITerminal from os.Stdout/os.Stdin
    - Enter raw mode
    - Enter alternate screen
    - Hide cursor
    - Get terminal size, create buffer
    - Create EventReader from os.Stdin
    - Create empty FocusManager
  - Implement `App.Close() error`
    - Show cursor
    - Exit alternate screen
    - Exit raw mode
    - Close EventReader
  - Implement `App.SetRoot(root Renderable)` (uses Renderable interface to avoid import cycle)
  - Implement `App.Root() Renderable`
  - Implement `App.Size() (width, height int)` from terminal
  - Implement `App.Focus() *FocusManager`
  - Implement `App.PollEvent(timeout time.Duration) (Event, bool)` wrapper
  - Implement `App.Dispatch(event Event) bool`
    - For `ResizeEvent`: resize buffer, mark root dirty, return true
    - For other events: delegate to FocusManager.Dispatch
  - Implement `App.Render()`
    - Clear buffer
    - If root exists: `root.Render(buf, width, height)`
    - Render to terminal via `Render(terminal, buffer)`
  - See [design §4](./phase5-event-system-design.md#4-app-type)

- [x] Create `pkg/tui/app_test.go`
  - Test `NewApp()` and `Close()` lifecycle (with mock terminal if needed)
  - Test `App.SetRoot()` and `App.Root()`
  - Test `App.Focus()` returns FocusManager
  - Test `App.Dispatch()` handles ResizeEvent
  - Test `App.Dispatch()` delegates KeyEvent to FocusManager
  - Test `App.Render()` integrates with element tree

- [x] Update `examples/dashboard/main.go` to use new event system
  - Replace manual goroutine-based input with `app.PollEvent()`
  - Use `App.Render()` instead of manual buffer management
  - Demonstrate proper cleanup with `defer app.Close()`
  - Keep animation loop structure but use polling pattern
  - See [design §5.3](./phase5-event-system-design.md#53-beforeafter-comparison)

- [x] Create `examples/focus/main.go` (new example demonstrating focus)
  - Create simple focusable elements (colored boxes that highlight on focus)
  - Register elements with FocusManager
  - Handle Tab/Shift+Tab for manual focus navigation
  - Handle Escape to exit
  - Demonstrate Focus/Blur visual feedback
  - See [design §5.1](./phase5-event-system-design.md#51-creating-a-focusable-element)

- [x] Create integration test `pkg/tui/integration_event_test.go`
  - Test full flow: MockEventReader → FocusManager → mock Focusable
  - Test resize event updates buffer dimensions
  - Test event dispatch to multiple focusable elements
  - Test focus cycling through registered elements

**Tests:** Run `go test ./pkg/tui/... -run "TestApp|TestIntegration"` once at phase end

---

## Phase Summary

| Phase | Description | Status |
|-------|-------------|--------|
| 1 | Core Event Types and Key Parsing | Complete |
| 2 | EventReader and FocusManager | Complete |
| 3 | App Integration and Examples | Complete |

## Files to Create

```
pkg/tui/
├── key.go              # Key and Modifier types
├── key_test.go
├── event.go            # Event interface and types
├── event_test.go
├── parse.go            # Escape sequence parser
├── parse_test.go
├── reader.go           # EventReader interface and stdinReader
├── reader_unix.go      # Unix-specific helpers (build-tagged)
├── reader_test.go
├── focus.go            # Focusable interface and FocusManager
├── focus_test.go
├── mock_reader.go      # MockEventReader for testing
├── app.go              # App type
├── app_test.go
└── integration_event_test.go

examples/
├── dashboard/main.go   # Updated to use event system
└── focus/main.go       # New focus demonstration
```

## Files to Modify

| File | Changes |
|------|---------|
| `examples/dashboard/main.go` | Rewrite to use `App` and `PollEvent` pattern |
