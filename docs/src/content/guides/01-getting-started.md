# Getting Started

## What is go-tui

go-tui is a declarative terminal UI framework for Go. You write UI layouts in `.gsx` files using a templ-like syntax with HTML-style elements, and the framework handles flexbox layout, rendering, and terminal input. It's pure Go with minimal external dependencies and generates type-safe code from your templates.

## Installation

First, add go-tui to your Go module:

```bash
go get github.com/grindlemire/go-tui
```

Then install the `tui` CLI tool, which compiles `.gsx` files into Go code:

```bash
go install github.com/grindlemire/go-tui/cmd/tui@latest
```

Make sure `$GOPATH/bin` (or `$GOBIN`) is in your `PATH` so the `tui` command is available.

## Editor Setup

For VS Code (or compatible editors like Cursor), install the official extension. It provides LSP support for `.gsx` files:

- [VS Code Marketplace](https://marketplace.visualstudio.com/items?itemName=grindlemire.go-tui)
- [Open VSX](https://open-vsx.org/extension/grindlemire/go-tui)

For Neovim and other editors, see the [Editor Integration](../reference/cli#editor-integration) section in the CLI reference.

## Your First App

This walks through building a "Hello, Terminal!" app from scratch.

### 1. Create the project

```bash
mkdir hello-tui && cd hello-tui
go mod init hello-tui
go get github.com/grindlemire/go-tui
```

### 2. Write hello.gsx

Create a file called `hello.gsx`:

```gsx
package main

import tui "github.com/grindlemire/go-tui"

type helloApp struct{}

func Hello() *helloApp {
	return &helloApp{}
}

func (h *helloApp) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnRune('q', func(ke tui.KeyEvent) { ke.App().Stop() }),
	}
}

templ (h *helloApp) Render() {
	<div class="flex-col items-center justify-center h-full">
		<div class="border-rounded border-cyan p-2 gap-1 flex-col items-center">
			<span class="text-cyan font-bold">Hello, Terminal!</span>
			<br />
			<span class="font-dim">Press q to quit</span>
		</div>
	</div>
}
```

This defines a struct component called `helloApp`. The `templ` keyword declares a `Render` method that returns a UI tree. Inside it, you use HTML-like elements (`<div>`, `<span>`, `<br />`) with Tailwind-style classes for layout and styling.

The outer `<div>` fills the full terminal height (`h-full`) and centers its children both horizontally (`items-center`) and vertically (`justify-center`). The inner `<div>` draws a rounded cyan border with padding and arranges its children in a column.

The `KeyMap()` method defines keyboard bindings. `tui.OnKey` matches special keys like Escape, and `tui.OnRune` matches printable characters. Each handler receives a `KeyEvent` with access to the app instance, so `ke.App().Stop()` exits the event loop and shuts down cleanly.

### 3. Write main.go

Create `main.go`:

```go
package main

import (
	"fmt"
	"os"

	tui "github.com/grindlemire/go-tui"
)

func main() {
	app, err := tui.NewApp(
		tui.WithRootComponent(Hello()),
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

`tui.NewApp` creates the application and puts the terminal into raw mode on an alternate screen. `tui.WithRootComponent` tells it which component to render. `app.Run()` starts the event loop and blocks until the app stops, and `app.Close()` restores the terminal to its original state.

### 4. Generate and run

Compile the `.gsx` file, then run the app:

```bash
tui generate hello.gsx
go run .
```

You should see a centered box with "Hello, Terminal!" in cyan. Press `q` or `Escape` to exit.

![Getting Started screenshot](/guides/01.png)

## How It Works

You write `.gsx` files using templ-like syntax, then `tui generate` compiles them into `_gsx.go` files containing standard Go code that calls the `tui` package API. From there, `go build` produces a single binary with no runtime dependencies. At runtime, the framework handles the event loop, flexbox layout, and double-buffered terminal rendering.

The generated `_gsx.go` files are recreated every time you run `tui generate` and should not be edited by hand.

## Core Concepts

**Components** come in two flavors. *Pure components* (`templ Greeting(name string) { ... }`) are stateless functions that take parameters and return UI. *Struct components* carry their own state, handle input, and support lifecycle hooks. See [GSX Syntax](gsx-syntax) and [Components](components).

**Elements** are the HTML-like tags you use in `.gsx` files: `<div>` for block containers, `<span>` for inline text, `<input />` for text fields, `<progress />` for progress bars, and more. See [GSX Syntax](gsx-syntax).

**Styling** uses Tailwind-inspired classes in the `class` attribute. Apply text colors (`text-cyan`), font styles (`font-bold`), borders (`border-rounded`), backgrounds (`bg-red`), and gradients (`text-gradient-cyan-magenta`). See [Styling and Colors](styling).

**Layout** follows the CSS flexbox model. Every `<div>` is a flex container. Control direction (`flex-col`), alignment (`items-center`, `justify-between`), spacing (`gap-2`, `p-1`), and sizing (`w-full`, `h-full`, `grow`) through classes or attributes. See [Layout](layout).

**State** is managed with the generic `State[T]` type. Create it with `tui.NewState(initialValue)`, read with `.Get()`, write with `.Set()` or `.Update()`. When state changes, the UI re-renders automatically. See [State and Reactivity](state).

**Events** cover keyboard and mouse input. Implement `KeyMap()` for key bindings (as shown above) or `HandleMouse()` for clicks and scrolling. See [Event Handling](events).

## Next Steps

- [GSX Syntax](gsx-syntax) — The full `.gsx` file format: elements, attributes, control flow, and code generation
- [Styling and Colors](styling) — Text styles, colors, borders, and gradients
