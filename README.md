<p align="center">
  <picture>
    <source media="(prefers-color-scheme: light)" srcset="docs/logos/final/go-tui-logo-light-bg.svg">
    <source media="(prefers-color-scheme: dark)" srcset="docs/logos/final/go-tui-logo.svg">
    <img alt="go-tui" src="docs/logos/final/go-tui-logo.svg" width="310">
  </picture>
</p>

<p align="center">
  <strong>Reactive Terminal UIs in Go</strong>
</p>

<p align="center">
  A Go framework for terminal UIs. Define layout in <code>.gsx</code> templates with HTML-like syntax, compile to type-safe Go. Flexbox positioning, reactive state, no CGO.
</p>

---

## Install

```bash
go get github.com/grindlemire/go-tui
```

Install the CLI tool:

```bash
go install github.com/grindlemire/go-tui/cmd/tui@latest
```

## Quick Example

**dashboard.gsx**

```gsx
package dashboard

templ Dashboard() {
  <div class="flex-col h-full">
    <div class="border-single p-1">
      <span class="font-bold text-cyan">Dashboard</span>
    </div>
    <div class="flex grow gap-2 p-1">
      @Sidebar()
      @MainContent()
    </div>
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
  app, err := tui.NewApp(
    tui.WithRootComponent(Dashboard()),
  )
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

Generate and run:

```bash
tui generate ./...
go run .
```

## Features

- **Declarative .gsx syntax** — HTML-like elements with Tailwind-style classes, compiled to type-safe Go
- **Pure Go flexbox** — Row, column, justify, align, gap, padding, margin. No CGO
- **Reactive state** — Generic `State[T]` with automatic re-rendering on change
- **Component system** — Parameters, refs, keyboard/mouse events, watchers
- **Editor support** — Language server, formatter, and tree-sitter grammar for VS Code and Zed
- **Minimal dependencies** — Only `golang.org/x` standard libraries

## Documentation

Full docs, guide, and API reference: [go-tui docs site](https://grindlemire.github.io/go-tui)

## License

MIT
