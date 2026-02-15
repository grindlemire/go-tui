---
title: "Getting Started"
order: 1
---

## Installation

Install the go-tui CLI and library:

```shell
go install github.com/grindlemire/go-tui/cmd/tui@latest
```

## Your First App

Create a simple hello world app. First, create a `hello.gsx` file:

```gsx
package main

templ Hello(name string) {
    <div class="flex-col p-2 border-rounded">
        <span class="font-bold text-cyan">Hello, {name}!</span>
        <span class="font-dim">Welcome to go-tui</span>
    </div>
}
```

## Generate & Run

Generate the Go code and run your app:

```shell
# Generate Go code from .gsx
tui generate hello.gsx

# Create main.go
cat > main.go << 'EOF'
package main

import (
    "fmt"
    "os"
    tui "github.com/grindlemire/go-tui"
)

func main() {
    app, err := tui.NewApp()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    defer app.Close()

    app.SetRoot(Hello("World").Render())
    if err := app.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
EOF

go run .
```
