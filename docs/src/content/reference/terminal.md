# Terminal Reference

## Overview

The `Terminal` interface abstracts all communication with the user's terminal. It handles screen rendering, cursor management, raw mode, alternate screen buffers, and mouse event reporting. go-tui ships with two implementations: `ANSITerminal` for real terminals and `MockTerminal` for testing.

Most applications never interact with `Terminal` directly; the `App` type handles terminal setup and teardown. You'll use these types when you need custom terminal behavior, direct rendering, or test verification.

## Terminal Interface

```go
type Terminal interface {
    Size() (width, height int)
    Flush(changes []CellChange)
    Clear()
    ClearToEnd()
    SetCursor(x, y int)
    HideCursor()
    ShowCursor()
    EnterRawMode() error
    ExitRawMode() error
    EnterAltScreen()
    ExitAltScreen()
    EnableMouse()
    DisableMouse()
    Caps() Capabilities
    WriteDirect([]byte) (int, error)
}
```

### Size

```go
Size() (width, height int)
```

Returns the terminal dimensions in cells (columns and rows).

### Flush

```go
Flush(changes []CellChange)
```

Writes a batch of cell changes to the terminal. Changes should be in row-major order for best performance. The implementation optimizes cursor movement and only emits style escape codes when the style differs from the previous cell.

### Clear

```go
Clear()
```

Clears the entire terminal screen, including scrollback on real terminals.

### ClearToEnd

```go
ClearToEnd()
```

Clears from the current cursor position to the end of the screen.

### SetCursor

```go
SetCursor(x, y int)
```

Moves the cursor to the given position. Coordinates are 0-indexed, with (0, 0) at the top-left corner.

### HideCursor / ShowCursor

```go
HideCursor()
ShowCursor()
```

Controls cursor visibility. `HideCursor` is typically called during rendering to prevent flicker, and `ShowCursor` restores visibility afterward.

### EnterRawMode / ExitRawMode

```go
EnterRawMode() error
ExitRawMode() error
```

Raw mode disables terminal line buffering so the application can read individual keystrokes. `EnterRawMode` saves the current terminal state; `ExitRawMode` restores it. Both return an error if the terminal state cannot be changed.

### EnterAltScreen / ExitAltScreen

```go
EnterAltScreen()
ExitAltScreen()
```

The alternate screen buffer is a separate display area that preserves the user's original terminal content. Full-screen applications use it so that quitting restores the previous scrollback. Inline-mode applications skip it.

### EnableMouse / DisableMouse

```go
EnableMouse()
DisableMouse()
```

Enables or disables SGR mouse event reporting. When enabled, mouse clicks, drags, and wheel scrolls generate `MouseEvent` values through the `EventReader`. Always call `DisableMouse` before exiting to restore normal terminal behavior.

### Caps

```go
Caps() Capabilities
```

Returns the terminal's detected capabilities: color support level, Unicode, true color, and alternate screen availability. See [Capabilities](#capabilities) below.

### WriteDirect

```go
WriteDirect([]byte) (int, error)
```

Writes raw bytes to the terminal output, bypassing the cell-based rendering pipeline. Use this for escape sequences or raw content that the higher-level API does not cover.

## ANSITerminal

`ANSITerminal` is the production implementation of `Terminal`. It communicates with real terminal emulators through ANSI escape sequences.

### Creating an ANSITerminal

#### NewANSITerminal

```go
func NewANSITerminal(out io.Writer, in io.Reader) (*ANSITerminal, error)
```

Creates a terminal with auto-detected capabilities. The `out` parameter is typically `os.Stdout` and `in` is `os.Stdin`. File descriptors are extracted from the writers if they implement `*os.File`, which enables raw mode and terminal size queries.

```go
import (
    "os"
    tui "github.com/grindlemire/go-tui"
)

term, err := tui.NewANSITerminal(os.Stdout, os.Stdin)
if err != nil {
    // handle error
}
```

#### NewANSITerminalWithCaps

```go
func NewANSITerminalWithCaps(out io.Writer, in io.Reader, caps Capabilities) *ANSITerminal
```

Creates a terminal with explicit capabilities instead of auto-detecting them. Use this when you know the target environment or want to force specific behavior.

```go
caps := tui.Capabilities{
    Colors:    tui.ColorTrue,
    Unicode:   true,
    TrueColor: true,
    AltScreen: true,
}
term := tui.NewANSITerminalWithCaps(os.Stdout, os.Stdin, caps)
```

### Additional Methods

`ANSITerminal` provides methods beyond the `Terminal` interface:

#### SetCaps

```go
func (t *ANSITerminal) SetCaps(caps Capabilities)
```

Updates the terminal's capabilities after creation. You might call this if you detect capabilities at runtime through a terminal query/response sequence.

#### ResetStyle

```go
func (t *ANSITerminal) ResetStyle()
```

Resets internal style tracking, forcing the next `Flush` call to emit style escape codes regardless of whether the style changed. Call this after writing directly to the terminal output to re-sync the style state.

#### Writer

```go
func (t *ANSITerminal) Writer() io.Writer
```

Returns the underlying output writer. Be careful: writes through this bypass the terminal's style optimization and cursor tracking.

#### BeginSyncUpdate / EndSyncUpdate

```go
func (t *ANSITerminal) BeginSyncUpdate()
func (t *ANSITerminal) EndSyncUpdate()
```

Brackets a synchronized update. Terminal emulators that support this protocol buffer output between `Begin` and `End`, then display everything at once. This prevents tearing during complex frame updates. Terminals that don't recognize the sequence just ignore it.

### Size Defaults

`ANSITerminal.Size()` returns 80x24 if the terminal dimensions cannot be determined (e.g., when stdout is not a TTY).

## BufferedWriter

A write buffer that batches output before sending it to an underlying writer.

```go
type BufferedWriter struct {
    // contains unexported fields
}
```

### NewBufferedWriter

```go
func NewBufferedWriter(out io.Writer) *BufferedWriter
```

Creates a `BufferedWriter` wrapping the given writer.

### Write

```go
func (w *BufferedWriter) Write(p []byte) (int, error)
```

Appends bytes to the internal buffer. Does not write to the underlying writer.

### Flush

```go
func (w *BufferedWriter) Flush() error
```

Writes the buffered content to the underlying writer and clears the buffer.

```go
bw := tui.NewBufferedWriter(os.Stdout)
bw.Write([]byte("hello "))
bw.Write([]byte("world"))
bw.Flush() // writes "hello world" to stdout in one call
```

## EventReader

The `EventReader` interface reads terminal input events for the application's event loop.

```go
type EventReader interface {
    PollEvent(timeout time.Duration) (Event, bool)
    Close() error
}
```

### PollEvent

```go
PollEvent(timeout time.Duration) (Event, bool)
```

Reads the next input event within the given timeout. Returns `(event, true)` when an event is available, or `(nil, false)` on timeout. Timeout behavior:

- **Positive duration**: waits up to that long for an event
- **Zero**: non-blocking check, returns immediately
- **Negative**: blocks indefinitely until an event arrives

Resize events (from SIGWINCH signals) are debounced with a 16ms window. Rapid resize signals during window dragging coalesce into a single event with the final dimensions.

### Close

```go
Close() error
```

Releases resources held by the reader (signal handlers, file descriptors). Must be called when the reader is no longer needed.

### NewEventReader

```go
func NewEventReader(in *os.File) (EventReader, error)
```

Creates an `EventReader` for the given input file (typically `os.Stdin`). The terminal should already be in raw mode before creating the reader. Sets up SIGWINCH handling for resize events on Unix systems.

```go
import (
    "os"
    tui "github.com/grindlemire/go-tui"
)

reader, err := tui.NewEventReader(os.Stdin)
if err != nil {
    // handle error
}
defer reader.Close()

event, ok := reader.PollEvent(100 * time.Millisecond)
if ok {
    // process event
}
```

## InterruptibleReader

Extends `EventReader` with the ability to wake up a blocking `PollEvent` call from another goroutine.

```go
type InterruptibleReader interface {
    EventReader

    EnableInterrupt() error
    Interrupt() error
}
```

### EnableInterrupt

```go
EnableInterrupt() error
```

Sets up the interrupt mechanism (a self-pipe on Unix). Must be called before using blocking mode (`PollEvent` with a negative timeout). Calling it multiple times is safe; subsequent calls are no-ops.

### Interrupt

```go
Interrupt() error
```

Wakes up a blocking `PollEvent` call. The blocked call returns `(nil, false)`. Safe to call even when `PollEvent` is not currently blocking.

The `App` uses `InterruptibleReader` internally to wake the event loop when state changes are queued via `QueueUpdate`.

## Capabilities

`Capabilities` describes what features the terminal supports.

```go
type Capabilities struct {
    Colors    ColorCapability
    Unicode   bool
    TrueColor bool
    AltScreen bool
}
```

| Field       | Type              | Description                                          |
|-------------|-------------------|------------------------------------------------------|
| `Colors`    | `ColorCapability` | Level of color support (none, 16, 256, or true color) |
| `Unicode`   | `bool`            | Whether the terminal renders Unicode characters       |
| `TrueColor` | `bool`            | Whether 24-bit RGB colors are supported               |
| `AltScreen` | `bool`            | Whether the alternate screen buffer is available      |

### DetectCapabilities

```go
func DetectCapabilities() Capabilities
```

Determines terminal capabilities from environment variables. Detection checks, in order:

1. `COLORTERM` — `truecolor` or `24bit` indicates true color support
2. Terminal-specific environment variables — `WT_SESSION` (Windows Terminal), `ITERM_SESSION_ID` (iTerm2), `KITTY_WINDOW_ID` (Kitty), `KONSOLE_VERSION` (Konsole), `VTE_VERSION` (GNOME Terminal, Tilix)
3. `TERM` — `dumb` disables all features; `256color` suffix enables 256-color; `truecolor` enables true color

Default capabilities when detection finds nothing specific: 16-color, Unicode enabled, no true color, alternate screen available.

```go
caps := tui.DetectCapabilities()
fmt.Println(caps) // e.g., "true-color, unicode, altscreen"
```

### ColorCapability Constants

```go
const (
    ColorNone ColorCapability = iota  // Monochrome, no color
    Color16                           // Standard 16 ANSI colors
    Color256                          // 256-color palette
    ColorTrue                         // 24-bit RGB true color
)
```

### SupportsColor

```go
func (c Capabilities) SupportsColor(color Color) bool
```

Reports whether the terminal can display the given color. ANSI colors require at least `Color16` support; RGB colors require `TrueColor`.

```go
caps := tui.DetectCapabilities()
color := tui.RGBColor(255, 128, 0) // orange

if caps.SupportsColor(color) {
    // use the color directly
} else {
    // fall back to an ANSI approximation
}
```

### EffectiveColor

```go
func (c Capabilities) EffectiveColor(color Color) Color
```

Returns the best color the terminal can actually display. If the terminal supports the color type, returns the original. Otherwise, RGB colors are approximated to the nearest ANSI color, and ANSI colors on colorless terminals fall back to `DefaultColor()`.

```go
caps := tui.DetectCapabilities()
requested := tui.RGBColor(0, 200, 100)
actual := caps.EffectiveColor(requested) // RGB on true-color, ANSI approximation otherwise
```

### String

```go
func (c Capabilities) String() string
```

Returns a human-readable summary like `"true-color, unicode, altscreen"` or `"16-color, ascii"`.

## Render Functions

Package-level functions that `App` uses internally for rendering. They're also available if you're building a custom render loop.

### Render

```go
func Render(term Terminal, buf *Buffer)
```

Diff-based render. Computes the cells that changed since the last `Swap()` and calls `term.Flush()` with only those changes. This is the standard render path and minimizes bytes sent to the terminal.

### RenderFull

```go
func RenderFull(term Terminal, buf *Buffer)
```

Full redraw. Sends every non-empty cell to the terminal regardless of whether it changed. Use this after operations that may have corrupted the terminal state (e.g., a child process writing to stdout).

## Cross-References

- [Testing Reference](testing.md) — `MockTerminal` and `MockEventReader` test utilities
- [Buffer Reference](buffer.md) — `Buffer`, `Cell`, and `CellChange` types used by `Flush`
- [Events Reference](events.md) — `Event`, `KeyEvent`, `MouseEvent` types returned by `PollEvent`
- [Styling Reference](styling.md) — `Style`, `Color`, and `Capabilities` color handling
- [App Reference](app.md) — `App` manages `Terminal` and `EventReader` lifecycle
