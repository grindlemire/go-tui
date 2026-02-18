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
  <a href="https://grindlemire.github.io/go-tui">Guides & API Reference</a> &middot;
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
- Language server, formatter, and tree-sitter grammar for VS Code and Zed
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

The [`examples/`](examples/) directory has runnable programs for each feature area:

| Example | What it covers |
|---------|----------------|
| [01-hello](examples/01-hello) | Minimal component, gradient text, quit handling |
| [02-styling](examples/02-styling) | Colors, text styles, borders, gradients, backgrounds |
| [03-layout](examples/03-layout) | Flexbox row/column, justify, align, gap, padding |
| [04-components](examples/04-components) | Standalone `templ` functions, `{children...}` slot |
| [05-state](examples/05-state) | `State[T]`, `@if`/`@for`/`@let`, reactive children |
| [06-keyboard](examples/06-keyboard) | `OnKey`, `OnRune`, special keys, Ctrl combos |
| [07-refs-and-clicks](examples/07-refs-and-clicks) | Refs, click handling, mouse + keyboard |
| [08-elements](examples/08-elements) | Built-in elements, disabled state, progress bars |
| [09-scrolling](examples/09-scrolling) | Scrollable containers, mouse wheel, PgUp/PgDn |
| [10-timers-and-watchers](examples/10-timers-and-watchers) | Interval timers, channel watchers, live data |
| [11-streaming](examples/11-streaming) | Auto-scroll, stick-to-bottom, streaming data |
| [12-multi-component](examples/12-multi-component) | Multi-file components, shared state |
| [13-dashboard](examples/13-dashboard) | Metrics, sparklines, scrollable event log |

```bash
cd examples/01-hello && go run .
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

**Zed** -- Extension at [`editor/zed-gsx/`](editor/zed-gsx/). Tree-sitter highlighting and LSP integration.

**Neovim / Helix** -- Tree-sitter grammar at [`editor/tree-sitter-gsx/`](editor/tree-sitter-gsx/).

The `tui lsp` language server works with any editor that speaks JSON-RPC over stdio. It proxies Go-specific features through gopls with `.gsx` ↔ `.go` source mapping.

## Documentation

Guides, syntax reference, and API docs: **[grindlemire.github.io/go-tui](https://grindlemire.github.io/go-tui)**

## License

MIT
