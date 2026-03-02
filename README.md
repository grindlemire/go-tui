<p align="center">
  <picture>
    <source media="(prefers-color-scheme: light)" srcset="docs/public/go-tui-logo-light-bg.svg">
    <source media="(prefers-color-scheme: dark)" srcset="docs/public/go-tui-logo.svg">
    <img alt="go-tui" src="docs/public/go-tui-logo.svg" width="310">
  </picture>
</p>

<p align="center">
  <strong>Reactive Terminal UIs in Go</strong>
</p>

<p align="center">
  Define terminal interfaces in <code>.gsx</code> templates with HTML-like syntax and Tailwind-style classes. <br>
  The compiler generates type-safe Go. The runtime handles flexbox layout, reactive state, and rendering.
</p>

<p align="center">
  <a href="https://go-tui.dev">Guides & API Reference</a> &middot;
  <a href="#examples">Examples</a> &middot;
  <a href="#editor-support">Editor Support</a>
</p>

---

## Install

```bash
go get github.com/grindlemire/go-tui
go install github.com/grindlemire/go-tui/cmd/tui@latest
```

## Quick look

**counter.gsx**

```go
package main

import (
    "fmt"
    tui "github.com/grindlemire/go-tui"
)

type counter struct {
    count *tui.State[int]
}

func Counter() *counter {
    return &counter{count: tui.NewState(0)}
}

func (c *counter) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.OnRune('+', func(ke tui.KeyEvent) { c.count.Update(func(v int) int { return v + 1 }) }),
        tui.OnRune('-', func(ke tui.KeyEvent) { c.count.Update(func(v int) int { return v - 1 }) }),
        tui.OnRune('q', func(ke tui.KeyEvent) { ke.App().Stop() }),
    }
}

templ (c *counter) Render() {
    <div class="flex-col items-center justify-center h-full gap-1">
        <span class="font-bold text-cyan">{fmt.Sprintf("Count: %d", c.count.Get())}</span>
        <span class="font-dim">+/- to change, q to quit</span>
    </div>
}
```

**main.go**

```go
package main

import (
    "fmt"
    "os"
    tui "github.com/grindlemire/go-tui"
)

func main() {
    app, err := tui.NewApp(tui.WithRootComponent(Counter()))
    if err != nil {
        fmt.Fprintf(os.Stderr, "%v\n", err)
        os.Exit(1)
    }
    defer app.Close()
    if err := app.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "%v\n", err)
        os.Exit(1)
    }
}
```

```bash
tui generate counter.gsx
go run .
```

The `generate` command compiles `.gsx` files into plain Go source (`*_gsx.go`) that you can read, debug, and commit.

## What's in the box

- `.gsx` templates with HTML-like elements and Tailwind-style utility classes, compiled to type-safe Go
- Pure Go flexbox layout -- row, column, justify, align, gap, padding, margin, percentage widths, min/max constraints, no CGO
- Generic `State[T]` with automatic re-rendering, batched updates, and bindings
- Struct components with keyboard/mouse handlers, watchers for timers and channels, refs, and a `{children...}` slot
- Language server, formatter, and tree-sitter grammar for VS Code
- Only depends on `golang.org/x/{sys,tools}` -- no vendored C, no termbox, no tcell

## How it works

```
.gsx files
    │  tui generate
    ▼
Go source (*_gsx.go)
    │  go build
    ▼
Widget tree + flexbox layout engine
    │
    ▼
Double-buffered character grid
    │  diff-based updates
    ▼
Terminal (ANSI escape sequences)
```

The `.gsx` compiler runs at build time and produces regular Go files. At runtime, the program builds a tree of `*tui.Element` nodes. The layout engine positions them with flexbox, and a double-buffered renderer diffs the output to minimize terminal writes.

## Examples

The [`examples/`](examples/) directory has runnable programs for each feature area. Examples 01–18 accompany the [guides](https://go-tui.dev).

| Example | What it covers |
|---------|----------------|
| [01-getting-started](examples/01-getting-started) | Minimal component, gradient text, quit handling |
| [02-gsx-syntax](examples/02-gsx-syntax) | GSX file structure, templ syntax, control flow |
| [03-styling](examples/03-styling) | Colors, text styles, borders, conditional styling |
| [04-layout](examples/04-layout) | Flexbox row/column, justify, align, reusable layouts |
| [05-elements](examples/05-elements) | Built-in elements, disabled state, progress bars |
| [06-state](examples/06-state) | `State[T]`, `@if`/`@for`/`@let`, reactive children |
| [07-components](examples/07-components) | Component composition, tabs, `{children...}` slot |
| [08-events](examples/08-events) | Keyboard event handling, `KeyMap`, `OnKey`/`OnRune` |
| [09-refs-and-clicks](examples/09-refs-and-clicks) | Refs, click handling, mouse + keyboard |
| [10-scrolling](examples/10-scrolling) | Scrollable containers, keyboard navigation |
| [11-focus](examples/11-focus) | Focus management, tab cycling |
| [12-watchers](examples/12-watchers) | Interval timers, channel watchers, live data |
| [13-testing](examples/13-testing) | Unit testing components |
| [14-multi-component](examples/14-multi-component) | Multi-file components, shared state |
| [15-inline-mode](examples/15-inline-mode) | Inline terminal rendering mode |
| [16-streaming](examples/16-streaming) | Auto-scroll, stick-to-bottom, streaming data |
| [17-inline-streaming](examples/17-inline-streaming) | Inline mode with streaming content |
| [18-dashboard](examples/18-dashboard) | Metrics, sparklines, scrollable event log |

See also [`ai-chat`](examples/ai-chat) and [`docs-example`](examples/docs-example).

```bash
cd examples/01-getting-started && go run .
```

## CLI

```
tui generate [path...]       Compile .gsx files to Go source
tui check [path...]          Validate without generating
tui fmt [path...]            Format .gsx files in place
tui fmt --check [path...]    Check formatting without modifying
tui lsp                      Start the language server (stdio)
tui version                  Print version
```

`tui generate ./...` compiles all `.gsx` files under the current directory.

## Editor support

**VS Code** -- Install the extension from [`editor/vscode/`](editor/vscode/). Syntax highlighting, completion, hover, go-to-definition, diagnostics, and formatting.

**Neovim / Helix** -- Tree-sitter grammar at [`editor/tree-sitter-gsx/`](editor/tree-sitter-gsx/).

The `tui lsp` language server works with any editor that speaks JSON-RPC over stdio. It proxies Go-specific features through gopls with `.gsx` ↔ `.go` source mapping.

## Documentation

Guides, syntax reference, and API docs: **[go-tui.dev](https://go-tui.dev)**

## Why are there so many files?

I wanted the imports to all be at the root so there is only a single package for the user to import and worry about (rather than an elements, styles, layout, etc.). Go Unfortunately doesn't have any indirection semantics for imports so I had to put a bunch of files in the root.

## License

MIT
