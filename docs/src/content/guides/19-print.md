# Single-Frame Printing

## Overview

Not every program needs an interactive event loop. Sometimes you want to render a styled table, a progress summary, or a formatted report once and exit. go-tui's `Print`, `Sprint`, and `Fprint` functions do exactly that: take a `Viewable`, run the flexbox layout engine, and emit ANSI-styled text to a writer. Same `.gsx` components, no `App` required.

## Quick Start

Write a component in a `.gsx` file just like you normally would:

```gsx
package main

import "fmt"

templ BuildReport(project string, status string, tests int, passed int) {
    <div class="flex-col border-rounded border-cyan p-1">
        <div class="flex-row justify-between">
            <span class="font-bold text-cyan">{project}</span>
            <span class="font-bold text-green">PASS</span>
        </div>
        <hr />
        <span class="text-dim">Tests: {fmt.Sprintf("%d/%d passed", passed, tests)}</span>
    </div>
}
```

Generate the Go code:

```bash
tui generate report.gsx
```

Then call `tui.Print` from your `main()`:

```go
package main

import tui "github.com/grindlemire/go-tui"

//go:generate go run ../../cmd/tui generate report.gsx

func main() {
    tui.Print(BuildReport("myapp", "PASS", 42, 42))
}
```

Run it and you get styled, bordered output printed to your terminal. The process exits immediately — no raw mode, no alternate screen, no event loop.

## Width Control

By default, `Print` detects the terminal width automatically (falling back to 80 columns if detection fails, for example when piping to a file). Override with `WithPrintWidth`:

```go
// Always render at 120 columns, regardless of terminal size
tui.Print(el, tui.WithPrintWidth(120))
```

This is useful when:
- Piping output to a file or another process
- Generating fixed-width output for CI logs
- Testing with deterministic widths

## Sprint and Fprint

`Sprint` returns the ANSI string instead of writing it:

```go
s := tui.Sprint(el, tui.WithPrintWidth(80))
fmt.Println(s) // you control where it goes
```

`Fprint` writes to any `io.Writer`:

```go
var buf bytes.Buffer
tui.Fprint(&buf, el, tui.WithPrintWidth(80))
// buf now contains the ANSI-styled output
```

```go
f, _ := os.Create("report.txt")
defer f.Close()
tui.Fprint(f, el, tui.WithPrintWidth(80))
```

Note: `Fprint` appends a trailing newline so the shell prompt doesn't collide with the output. `Sprint` does not, giving you full control when composing strings.

## Using the Same Components Everywhere

The `.gsx` components you write for single-frame printing also work with `App.Run()`. A component library can serve both interactive and non-interactive use cases:

```go
// Non-interactive: print and exit
tui.Print(StatusTable(results))

// Interactive: full TUI app
app, _ := tui.NewApp(tui.WithRootComponent(Dashboard(results)))
defer app.Close()
app.Run()
```

Both call the same generated functions. The only difference is how the element tree is consumed.
