// Package main demonstrates using reactive state bindings with DSL components.
// The counter value updates automatically when state changes - no manual
// SetText() calls needed.
//
// Run `go generate` to regenerate counter_tui.go from counter.tui.
//
// To build and run:
//
//	cd examples/counter-state
//	go run ../../cmd/tui generate counter.tui
//	go run .
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/grindlemire/go-tui/pkg/debug"
	"github.com/grindlemire/go-tui/pkg/layout"
	"github.com/grindlemire/go-tui/pkg/tui"
	"github.com/grindlemire/go-tui/pkg/tui/element"
)

//go:generate go run ../../cmd/tui generate counter.tui

func main() {
	// Create the application
	app, err := tui.NewApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create app: %v\n", err)
		os.Exit(1)
	}
	defer app.Close()

	// Build initial UI using generated component
	// Note: State is now internal to the component - no need to pass it
	root := buildUI(app)
	app.SetRoot(root)

	// Main event loop
	debug.Log("Starting main event loop")
	for {
		event, ok := app.PollEvent(50 * time.Millisecond)
		if ok {
			debug.Log("Received event: %T %+v", event, event)
			switch e := event.(type) {
			case tui.KeyEvent:
				debug.Log("KeyEvent: Key=%d Rune=%c", e.Key, e.Rune)
				switch {
				case e.Key == tui.KeyEscape || e.Rune == 'q':
					debug.Log("Quit requested")
					return
				default:
					// Dispatch other key events to focused element
					debug.Log("Dispatching key event to focused element")
					consumed := app.Dispatch(event)
					debug.Log("Event consumed: %v", consumed)
				}
			case tui.ResizeEvent:
				debug.Log("ResizeEvent: %dx%d", e.Width, e.Height)
				// Rebuild on resize to get new dimensions
				root = buildUI(app)
				app.SetRoot(root)
			}
		}

		app.Render()
	}
}

// buildUI creates the UI tree using the DSL-generated CounterUI component.
func buildUI(app *tui.App) *element.Element {
	width, height := app.Size()

	// Wrap the generated component in a root container
	root := element.New(
		element.WithSize(width, height),
		element.WithDirection(layout.Column),
		element.WithJustify(layout.JustifyCenter),
		element.WithAlign(layout.AlignCenter),
	)

	// Add the generated counter UI - now returns a view struct with .Root
	counter := CounterUI()
	root.AddChild(counter.Root)

	return root
}
