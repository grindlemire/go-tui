---
title: First Project
description: Create a minimal go-tui app from scratch.
---

# First Project

## What you'll learn
- How to create a new `go-tui` project.
- How to write a minimal `.gsx` component.
- How to generate and run the app.

## Prerequisites
- Read [Install and Run](./install-and-run).
- Have `tui` installed and on your `PATH`.

## Steps
1. Create a folder and initialize a Go module.
2. Add `go-tui` as a dependency.
3. Add `main.go` and `app.gsx`.
4. Generate code and run the app.

## Example
```bash
mkdir hello-tui
cd hello-tui
go mod init example.com/hello-tui
go get github.com/grindlemire/go-tui@latest
```

`main.go`:
```go
package main

import (
	"fmt"
	"os"

	tui "github.com/grindlemire/go-tui"
)

func main() {
	app, err := tui.NewApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create app: %v\n", err)
		os.Exit(1)
	}
	defer app.Close()

	app.SetRootComponent(Home())

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "app error: %v\n", err)
		os.Exit(1)
	}
}
```

`app.gsx`:
```go
package main

import tui "github.com/grindlemire/go-tui"

type home struct{}

func Home() *home {
	return &home{}
}

func (h *home) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnRune('q', func(tui.KeyEvent) { tui.Stop() }),
		tui.OnKey(tui.KeyEscape, func(tui.KeyEvent) { tui.Stop() }),
	}
}

templ (h *home) Render() {
	<div class="flex-col items-center justify-center h-full">
		<span class="font-bold text-green">My first go-tui app</span>
		<span class="font-dim">Press q to quit</span>
	</div>
}
```

Generate and run:
```bash
tui generate app.gsx
go run .
```

## Recap
You created a project, defined a component, generated Go code, and ran the app.

## Next
[Mental model](../core-concepts/mental-model)
