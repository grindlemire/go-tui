# Testing

## Overview

go-tui ships with `MockTerminal` and `MockEventReader` so you can test components without a real terminal. `MockTerminal` captures rendered output into an in-memory cell buffer you can inspect. `MockEventReader` queues up keyboard and mouse events and replays them on demand. Between them, you can render elements, simulate input, and assert on visual output and state changes from plain `go test`.

## MockTerminal

`MockTerminal` implements the `Terminal` interface. It maintains an internal grid of cells that records every write operation, so you can verify what your UI looks like after rendering.

Create one with a width and height (in characters and rows):

```go
term := tui.NewMockTerminal(80, 24)
```

### Reading Output

The two most common assertions use `String()` and `StringTrimmed()`:

```go
// Full buffer as a string (rows joined by newlines)
output := term.String()

// Same, but trailing spaces stripped from each line
output := term.StringTrimmed()
```

For cell-level checks, use `CellAt(x, y)`:

```go
cell := term.CellAt(5, 2)
if cell.Rune != 'H' {
    t.Errorf("CellAt(5, 2) = %q, want 'H'", cell.Rune)
}

// Check styling too
if !cell.Style.HasAttr(tui.AttrBold) {
    t.Error("expected bold text")
}
if !cell.Style.Fg.Equal(tui.Red) {
    t.Error("expected red foreground")
}
```

### Checking Terminal State

`MockTerminal` tracks mode switches so you can verify your app entered the right states:

```go
term.IsInRawMode()    // true after EnterRawMode()
term.IsInAltScreen()  // true after EnterAltScreen()
term.IsMouseEnabled() // true after EnableMouse()
term.IsCursorHidden() // true after HideCursor()
```

It also counts alternate screen transitions:

```go
term.AltScreenEnterCount() // how many times EnterAltScreen was called
term.AltScreenExitCount()  // how many times ExitAltScreen was called
```

### Resizing and Resetting

Resize the mock terminal to test layout changes:

```go
term.Resize(120, 40)
```

Reset clears the entire buffer and all state flags back to initial values:

```go
term.Reset()
```

You can also set terminal capabilities to test color degradation:

```go
term.SetCaps(tui.Capabilities{
    Colors:    tui.Color256,
    TrueColor: false,
    Unicode:   true,
    AltScreen: true,
})
```

## MockEventReader

`MockEventReader` implements `EventReader` and `InterruptibleReader`. It holds a queue of events and returns them one at a time through `PollEvent`.

Create one with events pre-loaded:

```go
reader := tui.NewMockEventReader(
    tui.KeyEvent{Key: tui.KeyRune, Rune: 'h'},
    tui.KeyEvent{Key: tui.KeyRune, Rune: 'i'},
    tui.KeyEvent{Key: tui.KeyEnter},
)
```

Pull events out with `PollEvent`. When the queue is empty, it returns `(nil, false)`:

```go
event, ok := reader.PollEvent(0)
if !ok {
    // no more events
}
```

Add more events after creation:

```go
reader.AddEvents(
    tui.KeyEvent{Key: tui.KeyEscape},
)
```

Check how many events are left:

```go
remaining := reader.Remaining()
```

Reset the reader to replay from the beginning:

```go
reader.Reset()
```

## Testing Rendered Output

The simplest test renders an element tree to a buffer and checks the result. No app setup needed.

```go
func TestPanel_RendersBorder(t *testing.T) {
    // Build the element tree
    root := tui.New(
        tui.WithSize(80, 24),
        tui.WithDirection(tui.Column),
    )

    panel := tui.New(
        tui.WithSize(20, 5),
        tui.WithBorder(tui.BorderSingle),
    )
    root.AddChild(panel)

    // Render to a buffer
    buf := tui.NewBuffer(80, 24)
    root.Render(buf, 80, 24)

    // Check the panel got the right dimensions
    rect := panel.Rect()
    if rect.Width != 20 || rect.Height != 5 {
        t.Errorf("panel size = %dx%d, want 20x5", rect.Width, rect.Height)
    }

    // Verify the top-left corner of the single border
    cell := buf.Cell(rect.X, rect.Y)
    if cell.Rune != '┌' {
        t.Errorf("top-left = %q, want '┌'", cell.Rune)
    }
}
```

For snapshot-style tests, render to a `MockTerminal` and check `StringTrimmed()`:

```go
func TestGreeting_ShowsMessage(t *testing.T) {
    buf := tui.NewBuffer(30, 5)
    term := tui.NewMockTerminal(30, 5)
    style := tui.NewStyle()

    buf.SetString(2, 1, "Hello, World!", style)
    tui.Render(term, buf)

    output := term.StringTrimmed()
    if !strings.Contains(output, "Hello, World!") {
        t.Errorf("expected 'Hello, World!' in output:\n%s", output)
    }
}
```

### Verifying Styled Output

You can check that cells carry the right styles by inspecting individual cell attributes:

```go
func TestStyledBox(t *testing.T) {
    buf := tui.NewBuffer(25, 7)
    term := tui.NewMockTerminal(25, 7)

    // Draw a rounded box with blue border
    boxStyle := tui.NewStyle().Foreground(tui.Blue)
    tui.DrawBox(buf, tui.NewRect(1, 1, 20, 5), tui.BorderRounded, boxStyle)

    // Draw bold red text inside
    textStyle := tui.NewStyle().Bold().Foreground(tui.Red)
    buf.SetString(3, 3, "Alert!", textStyle)

    tui.Render(term, buf)

    // Verify border style
    corner := term.CellAt(1, 1)
    if corner.Rune != '╭' {
        t.Errorf("top-left = %q, want '╭'", corner.Rune)
    }
    if !corner.Style.Fg.Equal(tui.Blue) {
        t.Error("border should be blue")
    }

    // Verify text style
    cell := term.CellAt(3, 3)
    if !cell.Style.HasAttr(tui.AttrBold) {
        t.Error("text should be bold")
    }
}
```

## Testing Key Handling

To test that a component's `KeyMap` responds correctly, build the component, call `KeyMap()`, and check the returned bindings against specific events. Here's a pattern that tests state changes after key presses:

```go
func TestCounter_KeyMap(t *testing.T) {
    type tc struct {
        key       tui.KeyEvent
        wantCount int
    }

    tests := map[string]tc{
        "increment with +": {
            key:       tui.KeyEvent{Key: tui.KeyRune, Rune: '+'},
            wantCount: 1,
        },
        "decrement with -": {
            key:       tui.KeyEvent{Key: tui.KeyRune, Rune: '-'},
            wantCount: -1,
        },
    }

    for name, tt := range tests {
        t.Run(name, func(t *testing.T) {
            c := NewCounter() // your component constructor
            km := c.KeyMap()

            // Find the matching binding and call its handler
            for _, binding := range km {
                if binding.Pattern.AnyRune && tt.key.IsRune() {
                    binding.Handler(tt.key)
                    break
                }
                if binding.Pattern.Rune == tt.key.Rune && tt.key.IsRune() {
                    binding.Handler(tt.key)
                    break
                }
                if binding.Pattern.Key == tt.key.Key && !tt.key.IsRune() {
                    binding.Handler(tt.key)
                    break
                }
            }

            if c.count.Get() != tt.wantCount {
                t.Errorf("count = %d, want %d", c.count.Get(), tt.wantCount)
            }
        })
    }
}
```

### Testing with FocusManager

For more realistic input testing, wire up a `FocusManager` with a `MockEventReader`. The `FocusManager` routes events to whatever element has focus, so you can test the full input pipeline.

You'll need a type that implements the `Focusable` interface (`IsFocusable`, `HandleEvent`, `Focus`, `Blur`). Here's a minimal test helper:

```go
type testFocusable struct {
    focused   bool
    lastEvent tui.Event
    handled   bool
}

func (f *testFocusable) IsFocusable() bool              { return true }
func (f *testFocusable) Focus()                          { f.focused = true }
func (f *testFocusable) Blur()                           { f.focused = false }
func (f *testFocusable) HandleEvent(e tui.Event) bool    { f.lastEvent = e; return f.handled }
```

Then use it in a test:

```go
func TestInput_ReceivesEvents(t *testing.T) {
    events := []tui.Event{
        tui.KeyEvent{Key: tui.KeyRune, Rune: 'h'},
        tui.KeyEvent{Key: tui.KeyRune, Rune: 'i'},
        tui.KeyEvent{Key: tui.KeyEnter},
    }

    reader := tui.NewMockEventReader(events...)
    elem := &testFocusable{handled: true}

    fm := tui.NewFocusManager()
    fm.Register(elem)

    // Process all events
    for {
        event, ok := reader.PollEvent(0)
        if !ok {
            break
        }
        fm.Dispatch(event)
    }

    // Last event should be Enter
    last := elem.lastEvent.(tui.KeyEvent)
    if last.Key != tui.KeyEnter {
        t.Errorf("last event = %v, want KeyEnter", last.Key)
    }
}
```

## Testing Resize Behavior

Test that your layout adapts to terminal size changes by rendering at different dimensions:

```go
func TestLayout_AdaptsToResize(t *testing.T) {
    type tc struct {
        width      int
        height     int
        wantPanelW int
    }

    tests := map[string]tc{
        "narrow terminal": {
            width:      40,
            height:     24,
            wantPanelW: 40,
        },
        "wide terminal": {
            width:      120,
            height:     24,
            wantPanelW: 120,
        },
    }

    for name, tt := range tests {
        t.Run(name, func(t *testing.T) {
            root := tui.New(
                tui.WithDirection(tui.Column),
            )

            panel := tui.New(
                tui.WithFlexGrow(1),
                tui.WithHeight(10),
            )
            root.AddChild(panel)

            buf := tui.NewBuffer(tt.width, tt.height)
            root.Render(buf, tt.width, tt.height)

            rect := panel.Rect()
            if rect.Width != tt.wantPanelW {
                t.Errorf("panel width = %d, want %d", rect.Width, tt.wantPanelW)
            }
        })
    }
}
```

## Table-Driven Tests

go-tui uses table-driven tests throughout. Define the `tc` struct before the test map, use `map[string]tc` for named cases, and iterate with `t.Run`:

```go
func TestMyComponent(t *testing.T) {
    type tc struct {
        initialCount int
        keyPresses   []rune
        wantCount    int
    }

    tests := map[string]tc{
        "no keys pressed": {
            initialCount: 0,
            keyPresses:   nil,
            wantCount:    0,
        },
        "three increments": {
            initialCount: 0,
            keyPresses:   []rune{'+', '+', '+'},
            wantCount:    3,
        },
        "increment then decrement": {
            initialCount: 5,
            keyPresses:   []rune{'+', '-'},
            wantCount:    5,
        },
    }

    for name, tt := range tests {
        t.Run(name, func(t *testing.T) {
            c := NewCounter()
            c.count.Set(tt.initialCount)

            km := c.KeyMap()
            for _, r := range tt.keyPresses {
                event := tui.KeyEvent{Key: tui.KeyRune, Rune: r}
                for _, binding := range km {
                    if binding.Pattern.Rune == r {
                        binding.Handler(event)
                        break
                    }
                }
            }

            if c.count.Get() != tt.wantCount {
                t.Errorf("count = %d, want %d", c.count.Get(), tt.wantCount)
            }
        })
    }
}
```

The `tc` struct lives inside the test function. Each test case has a descriptive name that shows up in `go test -v` output. The pattern keeps individual test functions short and makes it simple to add new cases.

## Tips

**Test rendering and behavior separately.** Rendering tests verify layout and visual output. Behavior tests verify state changes from key events. Mixing them makes failures harder to diagnose.

**Use `StringTrimmed()` over `String()`.** `String()` preserves trailing spaces on every line, which makes string comparisons brittle. `StringTrimmed()` strips trailing whitespace, giving you cleaner snapshot comparisons.

**Check cells, not full strings, for precise assertions.** `CellAt(x, y)` lets you verify a specific character and its style without worrying about surrounding whitespace. Use `strings.Contains` on `StringTrimmed()` for coarser checks.

**Keep your MockTerminal the same size as your Buffer.** If they differ, the render may clip or produce unexpected results.

## Next Steps

- [Multi-Component Applications](multi-component) -- Building apps with multiple struct components and shared state
