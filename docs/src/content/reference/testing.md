# Testing Utilities Reference

## Overview

go-tui ships two test doubles that let you verify component behavior without a real terminal: `MockTerminal` and `MockEventReader`. They replace the two runtime dependencies every go-tui app needs: a screen to draw on and a source of user input.

`MockTerminal` implements the full `Terminal` interface. It keeps an in-memory cell grid you can inspect after rendering. `MockEventReader` implements both `EventReader` and `InterruptibleReader`, returning pre-queued events one at a time so your tests can simulate any input sequence deterministically.

```go
import tui "github.com/grindlemire/go-tui"

term := tui.NewMockTerminal(80, 24)
reader := tui.NewMockEventReader(
    tui.KeyEvent{Key: tui.KeyRune, Rune: 'q'},
)
```

## MockTerminal

A simulated terminal backed by a flat cell buffer. All `Terminal` methods operate in memory instead of touching the real TTY, so you can inspect cursor movements, mode transitions, and cell writes after the fact.

### NewMockTerminal

```go
func NewMockTerminal(width, height int) *MockTerminal
```

Creates a mock terminal with the given dimensions. All cells start as spaces with default styling. The default capabilities are:

| Field     | Default   |
|-----------|-----------|
| Colors    | Color256  |
| Unicode   | true      |
| TrueColor | true      |
| AltScreen | true      |

```go
term := tui.NewMockTerminal(80, 24)
w, h := term.Size() // 80, 24
```

### Terminal Interface Methods

`MockTerminal` implements every method on the `Terminal` interface:

#### Size

```go
func (m *MockTerminal) Size() (width, height int)
```

Returns the current dimensions. Changes after calling `Resize`.

#### Flush

```go
func (m *MockTerminal) Flush(changes []CellChange)
```

Applies cell changes to the internal buffer. Out-of-bounds changes are silently ignored.

```go
term := tui.NewMockTerminal(80, 24)
term.Flush([]tui.CellChange{
    {X: 0, Y: 0, Cell: tui.NewCell('H', tui.NewStyle())},
    {X: 1, Y: 0, Cell: tui.NewCell('i', tui.NewStyle())},
})
// term now has "Hi" at the top-left
```

#### Clear

```go
func (m *MockTerminal) Clear()
```

Resets every cell to a space with default styling and moves the cursor to (0, 0).

#### ClearToEnd

```go
func (m *MockTerminal) ClearToEnd()
```

Clears from the current cursor position to the end of the screen.

#### SetCursor, HideCursor, ShowCursor

```go
func (m *MockTerminal) SetCursor(x, y int)
func (m *MockTerminal) HideCursor()
func (m *MockTerminal) ShowCursor()
```

Track cursor position and visibility. Query with `Cursor()` and `IsCursorHidden()`.

#### EnterRawMode, ExitRawMode

```go
func (m *MockTerminal) EnterRawMode() error
func (m *MockTerminal) ExitRawMode() error
```

Toggle raw mode state. Always return nil. Query with `IsInRawMode()`.

#### EnterAltScreen, ExitAltScreen

```go
func (m *MockTerminal) EnterAltScreen()
func (m *MockTerminal) ExitAltScreen()
```

Toggle alternate screen state. Each call increments the corresponding transition counter. Query with `IsInAltScreen()`, `AltScreenEnterCount()`, and `AltScreenExitCount()`.

#### EnableMouse, DisableMouse

```go
func (m *MockTerminal) EnableMouse()
func (m *MockTerminal) DisableMouse()
```

Toggle mouse reporting state. Query with `IsMouseEnabled()`.

#### Caps

```go
func (m *MockTerminal) Caps() Capabilities
```

Returns the current capabilities. Change them with `SetCaps()`.

#### WriteDirect

```go
func (m *MockTerminal) WriteDirect(b []byte) (int, error)
```

No-op. Returns `len(b), nil`. Raw escape sequences aren't processed in the mock.

### Test Helper Methods

These methods exist only on `MockTerminal` (not on the `Terminal` interface) and are for assertions in test code.

#### CellAt

```go
func (m *MockTerminal) CellAt(x, y int) Cell
```

Returns the `Cell` at the given position. Returns an empty `Cell` (zero rune, zero width) if the coordinates are out of bounds.

```go
term := tui.NewMockTerminal(80, 24)
buf := tui.NewBuffer(80, 24)
buf.SetString(0, 0, "Hello", tui.NewStyle())
tui.Render(term, buf)

cell := term.CellAt(0, 0)
fmt.Println(cell.Rune)           // 'H'
fmt.Println(cell.Style.HasAttr(tui.AttrBold)) // false
```

#### String

```go
func (m *MockTerminal) String() string
```

Renders the entire buffer to a string. Each row becomes a line separated by `\n`. Continuation cells from wide characters are skipped. The result has no trailing newline after the last row.

```go
term := tui.NewMockTerminal(5, 2)
term.Flush([]tui.CellChange{
    {X: 0, Y: 0, Cell: tui.NewCell('A', tui.NewStyle())},
    {X: 1, Y: 0, Cell: tui.NewCell('B', tui.NewStyle())},
})
fmt.Println(term.String())
// AB
//
```

#### StringTrimmed

```go
func (m *MockTerminal) StringTrimmed() string
```

Like `String`, but trailing spaces are removed from each line. Usually the better choice for assertions since it ignores empty space that would otherwise cause false mismatches.

```go
term := tui.NewMockTerminal(10, 3)
term.Flush([]tui.CellChange{
    {X: 0, Y: 0, Cell: tui.NewCell('H', tui.NewStyle())},
    {X: 1, Y: 0, Cell: tui.NewCell('i', tui.NewStyle())},
    {X: 2, Y: 1, Cell: tui.NewCell('X', tui.NewStyle())},
})
fmt.Println(term.StringTrimmed())
// Hi
//   X
//
```

#### Cursor

```go
func (m *MockTerminal) Cursor() (x, y int)
```

Returns the current cursor position.

#### IsCursorHidden

```go
func (m *MockTerminal) IsCursorHidden() bool
```

Returns `true` if `HideCursor()` was called without a subsequent `ShowCursor()`.

#### IsInRawMode

```go
func (m *MockTerminal) IsInRawMode() bool
```

Returns `true` if `EnterRawMode()` was called without a subsequent `ExitRawMode()`.

#### IsInAltScreen

```go
func (m *MockTerminal) IsInAltScreen() bool
```

Returns `true` if the terminal is currently in alternate screen mode.

#### AltScreenEnterCount

```go
func (m *MockTerminal) AltScreenEnterCount() int
```

Returns the total number of times `EnterAltScreen()` has been called since creation or the last `Reset()`. Useful for verifying that a component enters the alternate screen exactly once.

#### AltScreenExitCount

```go
func (m *MockTerminal) AltScreenExitCount() int
```

Returns the total number of times `ExitAltScreen()` has been called since creation or the last `Reset()`.

#### IsMouseEnabled

```go
func (m *MockTerminal) IsMouseEnabled() bool
```

Returns `true` if `EnableMouse()` was called without a subsequent `DisableMouse()`.

#### SetCaps

```go
func (m *MockTerminal) SetCaps(caps Capabilities)
```

Overrides the terminal's capabilities. Use this to test how your components behave under different terminal configurations.

```go
term := tui.NewMockTerminal(80, 24)

// Simulate a 16-color terminal
term.SetCaps(tui.Capabilities{
    Colors:    tui.Color16,
    Unicode:   false,
    TrueColor: false,
    AltScreen: false,
})

caps := term.Caps()
fmt.Println(caps.Colors) // Color16
```

#### Reset

```go
func (m *MockTerminal) Reset()
```

Returns the mock terminal to its initial state: clears all cells, resets the cursor to (0, 0), shows the cursor, exits raw mode, exits alternate screen, disables mouse, and zeros out the transition counters.

```go
term := tui.NewMockTerminal(80, 24)
term.HideCursor()
term.EnterAltScreen()
term.Reset()

fmt.Println(term.IsCursorHidden())         // false
fmt.Println(term.IsInAltScreen())          // false
fmt.Println(term.AltScreenEnterCount())    // 0
```

#### Resize

```go
func (m *MockTerminal) Resize(width, height int)
```

Changes the terminal dimensions. Content that falls within both the old and new bounds is preserved. Newly exposed cells are filled with spaces.

```go
term := tui.NewMockTerminal(5, 5)
term.Flush([]tui.CellChange{
    {X: 2, Y: 2, Cell: tui.NewCell('X', tui.NewStyle())},
})
term.Resize(10, 10)

w, h := term.Size()          // 10, 10
fmt.Println(term.CellAt(2, 2).Rune) // 'X' — preserved
```

#### SetCell

```go
func (m *MockTerminal) SetCell(x, y int, c Cell)
```

Directly writes a cell into the buffer. Out-of-bounds coordinates are silently ignored. Useful for setting up initial terminal state before running a test.

```go
term := tui.NewMockTerminal(80, 24)
term.SetCell(0, 0, tui.NewCell('Z', tui.NewStyle().Bold()))
fmt.Println(term.CellAt(0, 0).Rune) // 'Z'
```

## MockEventReader

A deterministic event source for testing. You pre-load it with events and they come back in order, one per `PollEvent` call.

### NewMockEventReader

```go
func NewMockEventReader(events ...Event) *MockEventReader
```

Creates a reader pre-loaded with the given events. The events are returned in order by successive calls to `PollEvent`. When the queue is exhausted, `PollEvent` returns `(nil, false)`.

```go
reader := tui.NewMockEventReader(
    tui.KeyEvent{Key: tui.KeyRune, Rune: 'h'},
    tui.KeyEvent{Key: tui.KeyRune, Rune: 'i'},
    tui.KeyEvent{Key: tui.KeyEnter},
)
```

### PollEvent

```go
func (m *MockEventReader) PollEvent(timeout time.Duration) (Event, bool)
```

Returns the next queued event. The `timeout` parameter is ignored. Events are returned immediately. When all events have been consumed, returns `(nil, false)`.

```go
reader := tui.NewMockEventReader(
    tui.KeyEvent{Key: tui.KeyEscape},
)

ev, ok := reader.PollEvent(0)    // KeyEvent{Key: KeyEscape}, true
ev, ok = reader.PollEvent(0)     // nil, false
```

### Close

```go
func (m *MockEventReader) Close() error
```

No-op. Always returns nil.

### AddEvents

```go
func (m *MockEventReader) AddEvents(events ...Event)
```

Appends more events to the end of the queue. Call this mid-test to simulate additional user input arriving after the initial batch.

```go
reader := tui.NewMockEventReader(
    tui.KeyEvent{Key: tui.KeyRune, Rune: 'a'},
)
reader.AddEvents(
    tui.KeyEvent{Key: tui.KeyRune, Rune: 'b'},
    tui.KeyEvent{Key: tui.KeyEnter},
)
fmt.Println(reader.Remaining()) // 3
```

### Remaining

```go
func (m *MockEventReader) Remaining() int
```

Returns the number of events still in the queue that haven't been consumed by `PollEvent`.

### Reset

```go
func (m *MockEventReader) Reset()
```

Rewinds the reader to the beginning so all originally queued events (plus any added with `AddEvents`) are returned again from the start.

```go
reader := tui.NewMockEventReader(
    tui.KeyEvent{Key: tui.KeyRune, Rune: 'x'},
)
reader.PollEvent(0) // consumes 'x'
reader.Reset()
ev, _ := reader.PollEvent(0)
fmt.Println(ev.(tui.KeyEvent).Rune) // 'x'
```

### EnableInterrupt

```go
func (m *MockEventReader) EnableInterrupt() error
```

No-op. Always returns nil. Exists to satisfy the `InterruptibleReader` interface.

### Interrupt

```go
func (m *MockEventReader) Interrupt() error
```

No-op. Always returns nil. Exists to satisfy the `InterruptibleReader` interface.

## Common Testing Patterns

### Render and Assert

Create a buffer, render content into it, flush to a mock terminal, and check the result.

```go
func TestGreeting(t *testing.T) {
    buf := tui.NewBuffer(20, 5)
    term := tui.NewMockTerminal(20, 5)

    buf.SetString(0, 0, "Hello", tui.NewStyle())
    tui.Render(term, buf)

    for i, r := range "Hello" {
        cell := term.CellAt(i, 0)
        if cell.Rune != r {
            t.Errorf("CellAt(%d, 0).Rune = %q, want %q", i, cell.Rune, r)
        }
    }
}
```

### Check Styled Content

Verify both the characters and their styling after rendering.

```go
func TestStyledText(t *testing.T) {
    buf := tui.NewBuffer(20, 5)
    term := tui.NewMockTerminal(20, 5)
    style := tui.NewStyle().Bold().Foreground(tui.Red)

    buf.SetString(3, 2, "Alert", style)
    tui.Render(term, buf)

    cell := term.CellAt(3, 2)
    if cell.Rune != 'A' {
        t.Errorf("Rune = %q, want 'A'", cell.Rune)
    }
    if !cell.Style.HasAttr(tui.AttrBold) {
        t.Error("expected bold text")
    }
    if !cell.Style.Fg.Equal(tui.Red) {
        t.Error("expected red foreground")
    }
}
```

### Table-Driven Tests

go-tui follows a consistent table-driven test convention. Define the test case struct separately, use a `map[string]tc` for the cases, and iterate with `t.Run`.

```go
func TestMockTerminal_Size(t *testing.T) {
    type tc struct {
        width, height int
    }

    tests := map[string]tc{
        "standard 80x24": {width: 80, height: 24},
        "large terminal":  {width: 200, height: 60},
        "small terminal":  {width: 40, height: 10},
    }

    for name, tt := range tests {
        t.Run(name, func(t *testing.T) {
            m := tui.NewMockTerminal(tt.width, tt.height)
            w, h := m.Size()
            if w != tt.width || h != tt.height {
                t.Errorf("Size() = (%d, %d), want (%d, %d)", w, h, tt.width, tt.height)
            }
        })
    }
}
```

### Testing with Borders

Verify that border drawing produces the expected box characters.

```go
func TestBorderedBox(t *testing.T) {
    buf := tui.NewBuffer(20, 6)
    term := tui.NewMockTerminal(20, 6)

    tui.DrawBox(buf, tui.NewRect(2, 1, 15, 4), tui.BorderSingle, tui.NewStyle())
    tui.Render(term, buf)

    if term.CellAt(2, 1).Rune != '┌' {
        t.Errorf("top-left = %q, want '┌'", term.CellAt(2, 1).Rune)
    }
    if term.CellAt(16, 1).Rune != '┐' {
        t.Errorf("top-right = %q, want '┐'", term.CellAt(16, 1).Rune)
    }
    if term.CellAt(2, 4).Rune != '└' {
        t.Errorf("bottom-left = %q, want '└'", term.CellAt(2, 4).Rune)
    }
    if term.CellAt(16, 4).Rune != '┘' {
        t.Errorf("bottom-right = %q, want '┘'", term.CellAt(16, 4).Rune)
    }
}
```

### Simulating Key Events

Create an app with a mock event reader to test how your component reacts to keyboard input.

```go
reader := tui.NewMockEventReader(
    tui.KeyEvent{Key: tui.KeyEnter},
)
app, err := tui.NewAppWithReader(reader,
    tui.WithRootComponent(MyApp()),
)
if err != nil {
    t.Fatal(err)
}
```

### Testing Terminal Capabilities

Swap out the default capabilities to verify your component behaves correctly on limited terminals.

```go
func TestLimitedTerminal(t *testing.T) {
    term := tui.NewMockTerminal(80, 24)
    term.SetCaps(tui.Capabilities{
        Colors:    tui.Color16,
        Unicode:   false,
        TrueColor: false,
        AltScreen: false,
    })

    caps := term.Caps()
    if caps.TrueColor {
        t.Error("expected no true color support")
    }
}
```

### Wide Character Support

Test CJK and emoji characters that occupy two cells.

```go
func TestWideCharacter(t *testing.T) {
    term := tui.NewMockTerminal(10, 3)

    term.Flush([]tui.CellChange{
        {X: 0, Y: 0, Cell: tui.NewCellWithWidth('中', tui.NewStyle(), 2)},
        {X: 1, Y: 0, Cell: tui.NewCellWithWidth(0, tui.NewStyle(), 0)}, // continuation
    })

    cell := term.CellAt(0, 0)
    if cell.Rune != '中' {
        t.Errorf("Rune = %q, want '中'", cell.Rune)
    }
    if cell.Width != 2 {
        t.Errorf("Width = %d, want 2", cell.Width)
    }
    if !term.CellAt(1, 0).IsContinuation() {
        t.Error("cell at (1,0) should be a continuation")
    }
}
```

## Related

- [App Reference](app.md) — `NewApp`, `NewAppWithReader`, and app lifecycle
- [Events Reference](events.md) — `KeyEvent`, `MouseEvent`, and key constants
- [Buffer Reference](buffer.md) — `Buffer`, `Cell`, and rendering functions
- [Testing Guide](../guides/10-testing.md) — step-by-step guide to writing tests
