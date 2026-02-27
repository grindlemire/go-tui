# Inline Streaming

## Overview

Guide 15 covered `PrintAbove` for printing complete lines above an inline widget. That works well for discrete messages, but sometimes you need character-by-character output. Think LLM responses arriving token by token, or progress updates that build a line incrementally.

`StreamAbove()` gives you a `*StreamWriter` that streams text into the history region above the inline widget. The writer implements `io.WriteCloser` for raw bytes (including ANSI escape sequences), and adds `WriteStyled` and `WriteGradient` methods for styled output without manual escape sequences. You write to it from a goroutine, and the framework takes care of coordinating with `PrintAbove`, wrapping lines, and rendering.

This guide builds on the [Inline Mode Guide](inline-mode) and the [Streaming Data Guide](streaming).

## StreamAbove vs PrintAbove

`PrintAbove` takes a format string and prints it all at once. It's the right tool when you have a complete line of text ready to display.

`StreamAbove()` returns a writer. You write bytes to it over time, and they appear above the widget as you write them. When you're done, you close the writer, and the framework finalizes whatever partial line remains.

| | PrintAbove | StreamAbove |
|---|---|---|
| Input | Complete formatted string | Incremental text via `*StreamWriter` |
| Styling | Format string with manual ANSI | `WriteStyled`, `WriteGradient`, or raw ANSI via `Write` |
| Thread safety | Use `QueuePrintAbove` from goroutines | Writer is goroutine-safe by default |
| Best for | Chat messages, log lines, status updates | LLM token streaming, progressive output |
| Element insertion | Use `PrintAboveElement` | Use `WriteElement` on the writer |

## Getting a Writer

Call `StreamAbove()` on the app to get a `*StreamWriter`:

```go
w := app.StreamAbove()
```

The writer is goroutine-safe. Writes are queued onto the main event loop internally, so you can write from any goroutine without synchronization.

If the app is not in inline mode, `StreamAbove()` returns a no-op writer that silently discards everything. This means you can use the same component code regardless of whether the app runs inline or full-screen.

Only one stream writer is active at a time. If you call `StreamAbove()` while a previous writer is still open, the framework finalizes the previous writer's partial line and closes it before returning the new one.

## Writing and Closing

Write bytes to the writer however you like. Each write appends characters to the current line above the widget:

```go
go func() {
    w := app.StreamAbove()
    for _, token := range tokens {
        fmt.Fprint(w, token)
        time.Sleep(30 * time.Millisecond)
    }
    w.Close()
}()
```

Newlines in the written data start a new line, just as you'd expect. The framework handles line wrapping when text exceeds the terminal width.

Close the writer when you're done. Closing finalizes the current partial line and makes it a permanent row in the history region. If you forget, the partial line stays pending until the next `StreamAbove()` or `PrintAbove` call cleans it up.

## Styled Streaming

The `StreamWriter` provides `WriteStyled` and `WriteGradient` methods so you don't need to construct ANSI escape sequences manually.

### WriteStyled

`WriteStyled` wraps text in a style's ANSI prefix and a reset suffix. It uses the terminal's detected capabilities to pick the right color encoding:

```go
w := app.StreamAbove()
w.WriteStyled("error: ", tui.NewStyle().Bold().Foreground(tui.Red))
w.WriteStyled("something went wrong\n", tui.NewStyle())
w.Close()
```

### WriteGradient

`WriteGradient` writes each character with an interpolated gradient foreground color. The writer tracks column position internally, so you don't need to manage counters:

```go
w := app.StreamAbove()
grad := tui.NewGradient(tui.Cyan, tui.Magenta)
for _, r := range text {
    w.WriteGradient(string(r), grad)
    time.Sleep(30 * time.Millisecond)
}
w.Close()
```

An optional base style provides additional attributes (bold, italic, etc.) and background color. The gradient color replaces the foreground:

```go
w.WriteGradient("bold gradient", grad, tui.NewStyle().Bold())
```

### Raw ANSI Escapes

You can also use plain `Write` to pass raw bytes, including ANSI escape sequences, through to the terminal:

```go
fmt.Fprintf(w, "\033[38;2;%d;%d;%dm%c", r, g, b, char)
fmt.Fprint(w, "\033[0m")
```

The framework's ANSI-aware byte scanner recognizes escape sequences and won't count them toward line width. A multi-byte SGR sequence has zero display width, so wrapping still happens at the right column.

## Coordinating with QueueUpdate

The stream writer handles terminal output, but it doesn't know about your component's state. If you need to update state while streaming (for example, tracking progress or toggling a "streaming" indicator), use `QueueUpdate` to schedule state changes on the main event loop:

```go
go func() {
    w := app.StreamAbove()
    for i, char := range text {
        fmt.Fprintf(w, "%c", char)
        app.QueueUpdate(func() {
            progress.Set(i)
        })
        time.Sleep(30 * time.Millisecond)
    }
    w.Close()
    app.QueueUpdate(func() {
        streaming.Set(false)
    })
}()
```

`QueueUpdate` is necessary here because `State.Set()` must happen on the main event loop. The stream writer's own writes are already queued internally, but your state updates are not.

## Interaction with PrintAbove

If `PrintAbove` or `PrintAboveln` is called while a stream writer is active, the framework finalizes the stream's partial line first, then prints the new content on the next line. This prevents interleaving.

The sequence looks like:

1. `StreamAbove()` returns a writer
2. You write "Hello wor" to the writer
3. `PrintAboveln("Status: ok")` is called
4. The framework finalizes "Hello wor" as a complete line
5. "Status: ok" appears on the next line
6. Further writes to the (now-closed) writer return `io.ErrClosedPipe`

## Inserting Elements Mid-Stream

`StreamWriter.WriteElement` lets you insert a fully rendered element into the scrollback without closing and reopening the stream. The element is laid out at the terminal width, rendered to ANSI text, and inserted row by row. Any partial line is finalized first.

This is useful for chat-style interfaces where streamed text includes structured content like tables:

```go
go func() {
    w := app.StreamAbove()
    w.WriteStyled("Here's the data:\n", tui.NewStyle().Bold())
    w.WriteElement(DataTable(rows))
    w.Write([]byte("Let me know if you need more.\n"))
    w.Close()
}()
```

`WriteElement` accepts any `*Element`, including output from templ functions. The element is rendered once and baked into static text — it does not remain interactive.

You can also use `PrintAboveElement` directly on the app when you're not mid-stream:

```go
app.PrintAboveElement(DataTable(rows))
```

Or from a goroutine:

```go
app.QueuePrintAboveElement(DataTable(rows))
```

## Complete Example

This example creates an inline widget with a 3-row status bar. Pressing Enter streams a phrase above the widget, one character at a time, with a gradient color effect:

```go
package main

import (
    "fmt"
    "os"

    tui "github.com/grindlemire/go-tui"
)

func main() {
    app, err := tui.NewApp(
        tui.WithInlineHeight(3),
        tui.WithRootComponent(StreamDemo()),
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

The component uses a `State[bool]` to track whether a stream is in progress, and disables starting another stream until the current one finishes:

```gsx
package main

import (
    "time"
    tui "github.com/grindlemire/go-tui"
)

var gradient = tui.NewGradient(tui.BrightCyan, tui.BrightMagenta)

var phrases = []string{
    "The quick brown fox jumps over the lazy dog.",
    "Stars scattered across the midnight canvas, each one a whisper of ancient light.",
    "Line one of a multi-line message.\nLine two continues here.\nAnd line three wraps it up.",
    "Streaming text appears character by character, just like a real-time API response.",
}

type streamDemo struct {
    app       *tui.App
    phraseIdx int
    streaming *tui.State[bool]
}

func StreamDemo() *streamDemo {
    return &streamDemo{
        streaming: tui.NewState(false),
    }
}

func (s *streamDemo) streamPhrase() {
    if s.streaming.Get() {
        return
    }
    s.streaming.Set(true)

    text := phrases[s.phraseIdx%len(phrases)]
    s.phraseIdx++

    go func() {
        w := s.app.StreamAbove()
        for _, r := range text {
            w.WriteGradient(string(r), gradient)
            time.Sleep(30 * time.Millisecond)
        }
        w.Close()
        s.app.QueueUpdate(func() {
            s.streaming.Set(false)
        })
    }()
}

func (s *streamDemo) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.OnKeyStop(tui.KeyEnter, func(ke tui.KeyEvent) { s.streamPhrase() }),
        tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
        tui.OnKey(tui.KeyCtrlC, func(ke tui.KeyEvent) { ke.App().Stop() }),
    }
}

func (s *streamDemo) statusText() string {
    if s.streaming.Get() {
        return "streaming..."
    }
    return "Press Enter to stream a phrase  |  Esc to quit"
}

templ (s *streamDemo) Render() {
    <div class="border-rounded border-cyan items-center justify-center">
        <span class="text-cyan">{s.statusText()}</span>
    </div>
}
```

A few things to notice in the component code:

`streamPhrase()` checks `streaming.Get()` and returns early if a stream is already running, so you can't start two at once.

The actual writing happens in a goroutine so it doesn't block the event loop. The 30ms sleep between characters creates the typing effect.

`WriteGradient` handles per-character gradient coloring and column tracking internally. No need for manual ANSI escape sequences, column counters, or `QueueUpdate` calls for position tracking.

Newlines in the text are handled automatically by `WriteGradient` — the column resets and a new line starts.

After writing all characters, we close the writer and set `streaming` back to `false` through `QueueUpdate`.

Generate and run:

```bash
tui generate ./...
go run .
```

Press Enter repeatedly to stream different phrases. Each appears above the widget with a cyan-to-magenta gradient:

![Inline Streaming screenshot](/guides/17.png)

## Next Steps

- [Inline Mode Guide](inline-mode) for `PrintAbove`, dynamic height, and alternate screen
- [Streaming Data Guide](streaming) for channel-based streaming into scrollable containers
- [Building a Dashboard](dashboard) for combining multiple streaming patterns
