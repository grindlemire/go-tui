// Package main demonstrates the component model with broadcast key dispatch.
//
// This example shows:
//   - Struct components with constructors (MyApp, Sidebar, SearchInput)
//   - KeyMap-based key binding with broadcast and stop propagation
//   - Conditional key activation (search input captures keys when active)
//   - Shared state between components via *tui.State[T] pointers
//   - Component lifecycle with Mount() caching
//
// To build and run:
//
//	go run ../../cmd/tui generate sidebar.gsx search.gsx app.gsx
//	go run .
package main

import (
	"fmt"
	"os"

	tui "github.com/grindlemire/go-tui"
)

//go:generate go run ../../cmd/tui generate sidebar.gsx search.gsx app.gsx

func main() {
	app, err := tui.NewApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create app: %v\n", err)
		os.Exit(1)
	}
	defer app.Close()

	// Create and set the root component. Child components are mounted
	// automatically via tui.Mount() in each Render() call.
	app.SetRootComponent(MyApp())

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "App error: %v\n", err)
		os.Exit(1)
	}
}
