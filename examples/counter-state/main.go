// Package main demonstrates reactive state with a simple counter.
//
// This example shows:
//   - State[T] for reactive values that trigger re-renders on change
//   - KeyMap for keyboard bindings
//   - HandleClicks for ref-based mouse hit testing
//   - Refs for button click detection
//
// To build and run:
//
//	go run ../../cmd/tui generate counter.gsx
//	go run .
package main

import (
	"fmt"
	"os"

	tui "github.com/grindlemire/go-tui"
)

//go:generate go run ../../cmd/tui generate counter.gsx

func main() {
	app, err := tui.NewApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create app: %v\n", err)
		os.Exit(1)
	}
	defer app.Close()

	app.SetRootComponent(Counter())

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "App error: %v\n", err)
		os.Exit(1)
	}
}
