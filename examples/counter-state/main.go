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

	tui "github.com/grindlemire/go-tui"
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

	app.SetGlobalKeyHandler(func(e tui.KeyEvent) bool {
		if e.Rune == 'q' || e.Key == tui.KeyEscape {
			app.Stop()
			return true // Event consumed
		}
		return false // Pass to focused element
	})

	err = app.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "App error: %v\n", err)
		os.Exit(1)
	}
}

// buildUI creates the UI tree using the DSL-generated CounterUI component.
func buildUI(app *tui.App) *tui.Element {
	width, height := app.Size()

	// Wrap the generated component in a root container
	root := tui.New(
		tui.WithSize(width, height),
		tui.WithDirection(tui.Column),
		tui.WithJustify(tui.JustifyCenter),
		tui.WithAlign(tui.AlignCenter),
	)

	// Add the generated counter UI - now returns a view struct with .Root
	counter := CounterUI()
	root.AddChild(counter.Root)

	return root
}
