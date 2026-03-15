# App Reference

## Overview

The `App` type is the top-level container that manages terminal setup, the event loop, rendering, and the component tree. Every go-tui program creates an `App`, sets a root component, calls `Run()`, and defers `Close()`.

```go
package main

import (
    "fmt"
    "os"

    tui "github.com/grindlemire/go-tui"
)

func main() {
    app, err := tui.NewApp(
        tui.WithRootComponent(MyApp()),
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

## Creating an App

### NewApp

```go
func NewApp(opts ...AppOption) (*App, error)
```

Creates a new application. Sets the terminal to raw mode and alternate screen (unless inline mode is configured). Pass `AppOption` functions to configure behavior.

Mouse behavior defaults:
- **Full screen mode**: mouse events enabled
- **Inline mode**: mouse events disabled (preserves terminal scrollback)
- Use `WithMouse()` or `WithoutMouse()` to override

```go
app, err := tui.NewApp(
    tui.WithRootComponent(MyApp()),
    tui.WithFrameRate(30),
)
```

### NewAppWithReader

```go
func NewAppWithReader(reader EventReader, opts ...AppOption) (*App, error)
```

Creates an App with a custom `EventReader`, typically for testing. Accepts the same options as `NewApp`.

```go
reader := tui.NewMockEventReader(
    tui.KeyEvent{Key: tui.KeyEnter},
)
app, err := tui.NewAppWithReader(reader,
    tui.WithRootComponent(MyApp()),
)
```

## AppOption Functions

Options configure the App before the event loop starts. Each returns an `AppOption` (which is `func(*App) error`).

### WithRootComponent

```go
func WithRootComponent(component Component) AppOption
```

Sets a struct component as the root. This is the most common way to wire up your UI. The component's `Render` method is called each frame when state is dirty, and the framework manages its lifecycle (key handling, mouse, watchers, mounting).

```go
tui.NewApp(tui.WithRootComponent(MyApp()))
```

### WithRoot

```go
func WithRoot(root *Element) AppOption
```

Sets a raw `*Element` as the root. Use this when you're building the element tree manually rather than through a struct component.

### WithRootView

```go
func WithRootView(view Viewable) AppOption
```

Sets a `Viewable` as the root. A `Viewable` provides both a root element and a set of watchers. The framework starts the watchers automatically when the view is set.

### WithFrameRate

```go
func WithFrameRate(fps int) AppOption
```

Sets the target render frame rate. Default is 60 fps. Valid range: 1-240.

```go
tui.WithFrameRate(30) // 30 fps, ~33ms per frame
```

### WithInputLatency

```go
func WithInputLatency(d time.Duration) AppOption
```

Sets the polling timeout for the event reader. Default is `tui.InputLatencyBlocking`, which blocks until input arrives (zero syscalls while idle). Use a positive duration like `50 * time.Millisecond` for polling mode. A value of 0 is not allowed.

```go
tui.WithInputLatency(50 * time.Millisecond) // poll every 50ms instead of blocking
```

### WithEventQueueSize

```go
func WithEventQueueSize(size int) AppOption
```

Sets the capacity of the internal event queue buffer. Default is 256. Must be at least 1.

### WithGlobalKeyHandler

```go
func WithGlobalKeyHandler(fn func(KeyEvent) bool) AppOption
```

Sets a handler that runs before key events reach the component tree (legacy path). If the handler returns `true`, the event is consumed and not dispatched further.

When using the component model (struct components with `KeyMap()`), key dispatch goes through the dispatch table instead. Prefer `KeyMap()` on your components over global handlers.

```go
tui.WithGlobalKeyHandler(func(ke tui.KeyEvent) bool {
    if ke.Key == tui.KeyEscape {
        ke.App().Stop()
        return true
    }
    return false
})
```

### WithMouse

```go
func WithMouse() AppOption
```

Explicitly enables mouse event reporting. Use this to turn on mouse support in inline mode, where it's off by default.

### WithoutMouse

```go
func WithoutMouse() AppOption
```

Explicitly disables mouse event reporting. Use this to turn off mouse support in full screen mode, where it's on by default.

### WithCursor

```go
func WithCursor() AppOption
```

Keeps the cursor visible during app execution. By default, the cursor is hidden.

### WithInlineHeight

```go
func WithInlineHeight(rows int) AppOption
```

Enables inline widget mode. The app manages only the bottom N rows of the terminal, and normal terminal output is preserved above. Must be at least 1.

In inline mode:
- Alternate screen is not used, so terminal history is preserved
- Mouse events are disabled by default
- Use `PrintAbove()` / `PrintAboveln()` to print scrolling content above the widget

```go
tui.WithInlineHeight(5) // 5-row widget at the bottom of the terminal
```

### WithInlineStartupMode

```go
func WithInlineStartupMode(mode InlineStartupMode) AppOption
```

Configures how inline mode handles existing visible terminal content at startup. See [InlineStartupMode Constants](#inlinestartupmode-constants) for the available modes.

### WithOnSuspend

```go
func WithOnSuspend(fn func()) AppOption
```

Sets a callback that runs before the app suspends on Ctrl+Z. Use this to save state or pause timers before the process stops.

### WithOnResume

```go
func WithOnResume(fn func()) AppOption
```

Sets a callback that runs after the app resumes from suspension (when the user runs `fg`). Use this to restart timers or refresh stale data.

## Lifecycle Methods

### Run

```go
func (a *App) Run() error
```

Starts the main event loop. Blocks until `Stop()` is called or a SIGINT (Ctrl+C) is received. The loop processes input events, re-renders when state is dirty, and sleeps for the remaining frame budget to maintain a consistent frame rate.

```go
if err := app.Run(); err != nil {
    fmt.Fprintf(os.Stderr, "Error: %v\n", err)
    os.Exit(1)
}
```

### Stop

```go
func (a *App) Stop()
```

Signals the `Run` loop to exit gracefully. All watchers receive a stop signal and their goroutines exit. Safe to call multiple times (idempotent).

```go
tui.On(tui.KeyEscape, func(ke tui.KeyEvent) {
    ke.App().Stop()
})
```

### Close

```go
func (a *App) Close() error
```

Restores the terminal to its original state: disables mouse, shows cursor, exits alternate screen (or clears the inline region), exits raw mode, and closes the event reader. Must be called when the application exits.

```go
app, err := tui.NewApp(...)
if err != nil { ... }
defer app.Close()
```

## Root Management

### SetRoot

```go
func (a *App) SetRoot(root *Element)
```

Sets a raw `*Element` as the root for rendering. Clears any previous root component, discovers focusable elements and watchers in the new tree, and marks the app dirty.

### SetRootView

```go
func (a *App) SetRootView(view Viewable)
```

Sets a `Viewable` as the root. Calls `BindApp` if the view implements `AppBinder`, extracts the root element via `GetRoot()`, and starts watchers returned by `GetWatchers()`.

### SetRootComponent

```go
func (a *App) SetRootComponent(component Component)
```

Sets a struct component as the root. Calls `BindApp` if the component implements `AppBinder`, renders it to produce the element tree, and marks the app dirty. On subsequent frames, the component's `Render` method is called again whenever the app is dirty.

### Root

```go
func (a *App) Root() *Element
```

Returns the current root element.

## Rendering

### Render

```go
func (a *App) Render()
```

Clears the buffer, re-renders the component tree (calling the root component's `Render` method if one is set), and flushes changes to the terminal. Handles buffer resizing, inline mode offsets, and the mark-and-sweep cycle for mounted sub-components. Automatically performs a full redraw after a resize.

### RenderFull

```go
func (a *App) RenderFull()
```

Forces a complete redraw of the entire buffer to the terminal. Use this after events that may corrupt the terminal display.

### MarkDirty

```go
func (a *App) MarkDirty()
```

Marks the app as needing a render on the next frame. You rarely need to call this directly; `State.Set()` and `State.Update()` call it automatically.

### Buffer

```go
func (a *App) Buffer() *Buffer
```

Returns the underlying double-buffered character grid. For advanced use cases only.

### Size

```go
func (a *App) Size() (width, height int)
```

Returns the current terminal dimensions in columns and rows.

### SnapshotFrame

```go
func (a *App) SnapshotFrame() string
```

Returns the current frame as a trimmed string. Useful for debugging and testing.

## Event Handling

### Dispatch

```go
func (a *App) Dispatch(event Event) bool
```

Sends an event through the framework's dispatch pipeline. Returns `true` if the event was consumed.

- **ResizeEvent**: updates buffer size, marks root dirty, schedules a full redraw.
- **MouseEvent**: hit-tests the element tree to find the target element and dispatches to it.
- **KeyEvent**: routes through the focus manager to the focused element.

This method is primarily used in tests. During normal operation, `Run()` handles dispatch internally.

```go
app.Dispatch(tui.KeyEvent{Key: tui.KeyEnter})
```

### SetGlobalKeyHandler

```go
func (a *App) SetGlobalKeyHandler(fn func(KeyEvent) bool)
```

Sets (or replaces) a handler that runs before key events reach the focus manager. If the handler returns `true`, the event is consumed.

### QueueUpdate

```go
func (a *App) QueueUpdate(fn func())
```

Enqueues a function to run on the main event loop. Safe to call from any goroutine. Use this when you need to update state from a background goroutine.

```go
go func() {
    result := fetchData()
    app.QueueUpdate(func() {
        data.Set(result)
    })
}()
```

### PollEvent

```go
func (a *App) PollEvent(timeout time.Duration) (Event, bool)
```

Reads the next input event with a timeout. Returns the event and `true` if one was available, or a zero event and `false` on timeout. Convenience wrapper around the `EventReader`.

## Focus

### FocusNext

```go
func (a *App) FocusNext()
```

Moves focus to the next focusable element in document order.

### FocusPrev

```go
func (a *App) FocusPrev()
```

Moves focus to the previous focusable element.

### Focused

```go
func (a *App) Focused() Focusable
```

Returns the currently focused element, or `nil` if nothing is focused.

## Inline Mode

These methods only take effect when the app was created with `WithInlineHeight`. In full-screen mode they are no-ops.

### PrintAbove

```go
func (a *App) PrintAbove(format string, args ...any)
```

Prints formatted content above the inline widget. Does not add a trailing newline. Must be called from the app's main event loop.

### PrintAboveln

```go
func (a *App) PrintAboveln(format string, args ...any)
```

Same as `PrintAbove`, but appends a newline. Must be called from the main event loop.

### QueuePrintAbove

```go
func (a *App) QueuePrintAbove(format string, args ...any)
```

Thread-safe version of `PrintAbove`. Safe to call from any goroutine.

### QueuePrintAboveln

```go
func (a *App) QueuePrintAboveln(format string, args ...any)
```

Thread-safe version of `PrintAboveln`. Safe to call from any goroutine.

### StreamAbove

```go
func (a *App) StreamAbove() *StreamWriter
```

Returns a `*StreamWriter` that streams text character by character to the history region above the inline widget. The writer implements `io.WriteCloser` for plain byte streaming, and adds `WriteStyled` and `WriteGradient` methods for styled output. The writer is goroutine-safe: writes are queued onto the main event loop internally.

Closing the writer finalizes the current partial line, making it a permanent row in the history. If a previous stream writer is still open when `StreamAbove()` is called again, the framework finalizes and closes it before returning the new one.

Returns a no-op writer (silently discards all bytes) when not in inline mode.

```go
go func() {
    w := app.StreamAbove()
    // Plain write (backward compatible with io.WriteCloser)
    fmt.Fprint(w, "hello ")
    // Styled write
    w.WriteStyled("important", tui.NewStyle().Bold().Foreground(tui.Red))
    // Gradient write (per-character color interpolation)
    grad := tui.NewGradient(tui.Cyan, tui.Magenta)
    w.WriteGradient("gradient text", grad)
    // Gradient with base style attributes
    w.WriteGradient("bold gradient", grad, tui.NewStyle().Bold())
    w.Close()
}()
```

#### StreamWriter Methods

| Method | Signature | Description |
|--------|-----------|-------------|
| `Write` | `Write(p []byte) (int, error)` | Plain byte write (backward compatible). Does not track column position. |
| `Close` | `Close() error` | Finalizes the partial line. |
| `WriteStyled` | `WriteStyled(text string, style Style) (int, error)` | Writes text with ANSI style prefix and reset suffix. Tracks column position. |
| `WriteGradient` | `WriteGradient(text string, g Gradient, base ...Style) (int, error)` | Writes each character with an interpolated gradient foreground color. Optional base style provides attributes (bold, italic) and background. Tracks column position. |
| `WriteElement` | `WriteElement(el *Element)` | Renders an element and inserts its rows mid-stream. Finalizes the current partial line first. Resets column position. |

Column tracking: `WriteStyled` and `WriteGradient` advance an internal column counter by each character's display width (CJK-aware). Newlines reset the column to 0. The column wraps at the terminal width. This counter drives gradient color interpolation.

If `PrintAbove` or `PrintAboveln` is called while a stream writer is active, the stream's partial line is finalized first, then the new content is printed on the next line. Further writes to the finalized writer return `io.ErrClosedPipe`.

### PrintAboveElement

```go
func (a *App) PrintAboveElement(el *Element)
```

Renders an element tree and inserts the resulting rows into the inline scrollback as static ANSI text. The element is laid out at the terminal's current width using the standard flexbox layout engine, then each row is converted to ANSI escape sequences and inserted via the same machinery as `PrintAboveStyled`.

This enables inserting structured content (tables, styled cards, templ component output) into the scrollback alongside streamed text.

Must be called from the app's main event loop. No-op if not in inline mode or if the element is nil.

```go
// Insert a templ-generated table
table := DataTable(myRows)
app.PrintAboveElement(table)

// Insert a hand-built element
el := tui.NewElement(tui.WithText("result"), tui.WithBorder(tui.BorderSingle))
app.PrintAboveElement(el)
```

Any active stream writer's partial line is finalized before the element is inserted, same as `PrintAbove`.

### QueuePrintAboveElement

```go
func (a *App) QueuePrintAboveElement(el *Element)
```

Thread-safe version of `PrintAboveElement`. Safe to call from any goroutine. The element is rendered and inserted on the main event loop.

```go
go func() {
    table := buildTable(result)
    app.QueuePrintAboveElement(table)
}()
```

### SetInlineHeight

```go
func (a *App) SetInlineHeight(rows int)
```

Changes the inline widget height at runtime. The height change takes effect immediately. Capped at terminal height, minimum of 1. Should be called from render functions or the main event loop.

### InlineHeight

```go
func (a *App) InlineHeight() int
```

Returns the current inline height. Returns 0 if the app is not in inline mode.

## Alternate Screen

These methods let an inline-mode app switch to a full-screen overlay and back, for things like settings panels or help screens that shouldn't affect terminal scrollback.

### EnterAlternateScreen

```go
func (a *App) EnterAlternateScreen() error
```

Switches to alternate screen mode for a full-screen UI overlay. Saves the current inline mode state, clears the inline region, enters the alternate screen, and resizes the buffer to full terminal dimensions. No-op if already in alternate mode.

### ExitAlternateScreen

```go
func (a *App) ExitAlternateScreen() error
```

Returns from alternate screen mode and restores the previous inline mode state. The terminal content from before `EnterAlternateScreen` reappears. No-op if not in alternate mode.

### IsInAlternateScreen

```go
func (a *App) IsInAlternateScreen() bool
```

Returns `true` if the app is currently displaying in the alternate screen overlay.

## Job Control (Ctrl+Z)

Pressing Ctrl+Z suspends the app, and running `fg` in the shell resumes it. This works automatically for all go-tui apps.

On suspend, the framework automatically:
1. Disables mouse reporting
2. Shows the cursor
3. Exits the alternate screen (full-screen mode)
4. Restores normal terminal mode

On resume, it reverses all of these and forces a full redraw.

### Hooks

Use `WithOnSuspend` and `WithOnResume` to run custom logic:

```go
app, err := tui.NewApp(
    tui.WithOnSuspend(func() {
        // Save state before suspend
    }),
    tui.WithOnResume(func() {
        // Refresh data after resume
    }),
)
```

### Suspend

```go
func (a *App) Suspend()
```

Programmatically triggers a suspend (same as Ctrl+Z). Safe to call from any goroutine.

### Overriding Ctrl+Z

To prevent Ctrl+Z from suspending, bind `KeyCtrlZ` with a Stop handler in a component's `KeyMap()`:

```go
func (c *myComponent) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.OnStop(tui.KeyCtrlZ, func(ke tui.KeyEvent) {
            // Custom behavior instead of suspend
        }),
    }
}
```

## State Batching

### Batch

```go
func (a *App) Batch(fn func())
```

Executes `fn` with all state binding callbacks deferred until `fn` returns. Multiple `Set()` calls within a batch coalesce into a single round of binding execution, which avoids redundant intermediate renders.

```go
app.Batch(func() {
    firstName.Set("Alice")
    lastName.Set("Smith")
}) // Bindings fire once here, not twice
```

## Component Mounting

### Mount

```go
func (a *App) Mount(parent Component, index int, factory func() Component) *Element
```

Creates or retrieves a cached component instance and returns its rendered element tree. Generated code calls this; you typically don't call it directly.

On first call for a given `(parent, index)` pair: executes `factory`, caches the instance, calls `BindApp` and `Init()` (if implemented). On subsequent calls: returns the cached instance's `Render()` result. If the instance implements `PropsUpdater`, `UpdateProps` is called with a fresh instance so the cached component can pick up new props.

Stale cache entries are swept after each render pass. When a component disappears from the tree, its cleanup function runs.

## Other

### Terminal

```go
func (a *App) Terminal() Terminal
```

Returns the underlying `Terminal` implementation. For advanced use cases such as direct ANSI escape output.

### EventQueue

```go
func (a *App) EventQueue() chan<- func()
```

Returns the event queue channel for manual watcher setup. Prefer `WatcherProvider` on components instead.

### StopCh

```go
func (a *App) StopCh() <-chan struct{}
```

Returns a channel that closes when the app stops. For manual watcher setup. Prefer `WatcherProvider` on components instead.

## InlineStartupMode Constants

These constants control how inline mode initializes the visible terminal viewport at startup. Pass them to `WithInlineStartupMode`.

| Constant | Description |
|----------|-------------|
| `InlineStartupPreserveVisible` | Default. Keeps existing visible rows on launch. Unknown history drains naturally as new `PrintAbove` content is appended. |
| `InlineStartupFreshViewport` | Clears the visible viewport immediately. Existing visible rows are discarded. |
| `InlineStartupSoftReset` | Pushes existing visible rows into scrollback via newline flow, then clears the viewport. |

## InputLatencyBlocking Constant

```go
const InputLatencyBlocking = -1 * time.Millisecond
```

The default value for `WithInputLatency`. The event reader blocks until input arrives, producing zero syscalls while idle. The framework sets up the interrupt pipe automatically so that SIGWINCH and shutdown can wake the blocked reader.

## Interfaces

These interfaces are relevant when working with the App's root management methods.

### Viewable

```go
type Viewable interface {
    GetRoot() *Element
    GetWatchers() []Watcher
}
```

Implemented by `*Element`, generated view structs, and struct components. Allows `SetRootView` to extract the root element and start watchers. Also accepted by `PrintAboveElement` and `StreamWriter.WriteElement`.

### Component

```go
type Component interface {
    Render(app *App) *Element
}
```

The base interface for struct components. Any struct with a matching `Render` method can serve as a component.

### Optional Component Interfaces

Components can implement these additional interfaces for extended behavior:

| Interface | Method | Purpose |
|-----------|--------|---------|
| `KeyListener` | `KeyMap() KeyMap` | Keyboard handling |
| `MouseListener` | `HandleMouse(MouseEvent) bool` | Mouse handling |
| `WatcherProvider` | `Watchers() []Watcher` | Timers and channel watchers |
| `Initializer` | `Init() func()` | Setup on mount, returns cleanup function |
| `AppBinder` | `BindApp(app *App)` | Receives app reference (called automatically by mount) |
| `AppUnbinder` | `UnbindApp()` | Detaches app-bound resources on unmount |
| `PropsUpdater` | `UpdateProps(fresh Component)` | Receives updated props on re-render from cache |

See the [Component Interfaces Reference](interfaces.md) for full details on each.
