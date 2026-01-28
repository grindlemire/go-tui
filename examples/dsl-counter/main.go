// Package main demonstrates using DSL-generated components.
// Run `go generate` to regenerate counter_gsx.go from counter.gsx.
//
// To build and run:
//
//	cd examples/dsl-counter
//	go run ../../cmd/tui generate counter.gsx
//	go run .
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/grindlemire/go-tui/pkg/layout"
	"github.com/grindlemire/go-tui/pkg/tui"
	"github.com/grindlemire/go-tui/pkg/tui/element"
)

//go:generate go run ../../cmd/tui generate counter.gsx

func main() {
	// Create the application
	app, err := tui.NewApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create app: %v\n", err)
		os.Exit(1)
	}
	defer app.Close()

	// Counter state
	count := 0

	// Build initial UI using generated component
	root := buildUI(app, count)
	app.SetRoot(root)

	// Main event loop
	for {
		event, ok := app.PollEvent(50 * time.Millisecond)
		if ok {
			switch e := event.(type) {
			case tui.KeyEvent:
				switch {
				case e.Key == tui.KeyEscape || e.Rune == 'q':
					return
				case e.Rune == '+' || e.Rune == '=':
					count++
					root = buildUI(app, count)
					app.SetRoot(root)
				case e.Rune == '-' || e.Rune == '_':
					count--
					root = buildUI(app, count)
					app.SetRoot(root)
				}
			case tui.ResizeEvent:
				// Rebuild on resize to get new dimensions
				root = buildUI(app, count)
				app.SetRoot(root)
			}
		}

		app.Render()
	}
}

// buildUI creates the UI tree using the DSL-generated CounterUI component.
func buildUI(app *tui.App, count int) *element.Element {
	width, height := app.Size()

	// Wrap the generated component in a root container
	root := element.New(
		element.WithSize(width, height),
		element.WithDirection(layout.Column),
		element.WithJustify(layout.JustifyCenter),
		element.WithAlign(layout.AlignCenter),
	)

	// Add the generated counter UI - now returns a view struct with .Root
	counter := CounterUI(count)
	root.AddChild(counter.Root)

	return root
}
