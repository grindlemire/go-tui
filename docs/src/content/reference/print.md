# Print Reference

## Overview

The `Print`, `Sprint`, and `Fprint` functions render a `Viewable` once with full ANSI styling, without starting an interactive `App`. They run the flexbox layout engine, render into an internal buffer, and convert each row to ANSI-escaped text. Use them for CLI tools, reports, and any program that wants styled terminal output without an event loop.

All three functions accept any `Viewable`, including generated `.gsx` component views and raw `*Element` values.

## Functions

### Print

```go
func Print(v Viewable, opts ...PrintOption)
```

Renders a `Viewable` to `os.Stdout` with ANSI styling. Appends a trailing newline after the output. Width is auto-detected from the terminal, falling back to 80 columns.

```go
tui.Print(BuildReport("myapp", "PASS", "2.3s", 42, 42))
```

With explicit width:

```go
tui.Print(view, tui.WithPrintWidth(120))
```

### Sprint

```go
func Sprint(v Viewable, opts ...PrintOption) string
```

Renders a `Viewable` to a string containing ANSI escape codes. Does not append a trailing newline. Returns an empty string if the element has no renderable content.

```go
s := tui.Sprint(view, tui.WithPrintWidth(80))
fmt.Println(s)
```

### Fprint

```go
func Fprint(w io.Writer, v Viewable, opts ...PrintOption)
```

Renders a `Viewable` to the given `io.Writer` with ANSI styling. Appends a trailing newline after the output. Writes nothing if the element has no renderable content.

```go
var buf bytes.Buffer
tui.Fprint(&buf, view, tui.WithPrintWidth(80))
```

```go
f, _ := os.Create("output.txt")
defer f.Close()
tui.Fprint(f, view)
```

## Options

### PrintOption

```go
type PrintOption func(*printConfig)
```

Functional option type for configuring single-frame rendering.

### WithPrintWidth

```go
func WithPrintWidth(w int) PrintOption
```

Sets an explicit render width in characters. When not set, width is auto-detected from the terminal with a fallback to 80 columns.

| Parameter | Type | Description |
|-----------|------|-------------|
| `w` | `int` | Width in terminal columns |

```go
tui.Print(view, tui.WithPrintWidth(120))
```

## Width Detection

When no explicit width is provided, the functions query the terminal for its current width using the platform's terminal size API:

| Platform | Mechanism |
|----------|-----------|
| macOS / BSD | `unix.IoctlGetWinsize` via `TIOCGWINSZ` |
| Linux / Solaris | `unix.IoctlGetWinsize` via `TIOCGWINSZ` |
| Windows | `windows.GetConsoleScreenBufferInfo` |

If detection fails (for example, when stdout is piped to a file or another process), the width defaults to 80 columns.

## Capabilities

Color and style support is auto-detected via `DetectCapabilities()`, which reads these environment variables:

| Variable | Effect |
|----------|--------|
| `COLORTERM=truecolor` or `24bit` | Enables 24-bit true color |
| `TERM` containing `256color` | Enables 256-color palette |
| `TERM=dumb` | Disables all color and unicode |
| `WT_SESSION`, `ITERM_SESSION_ID`, `KITTY_WINDOW_ID` | Enables true color for known emulators |

RGB colors are automatically downgraded to the best available palette when true color is not supported.

## Height

Height is always determined by the element's natural content size. The element renders at whatever height the flexbox layout computes for the given width. There is no height override option.
