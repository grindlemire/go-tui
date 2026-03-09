// Package main demonstrates all built-in elements: input, button, ul/li, table, progress, p, hr.
//
// To build and run:
//
//	go run ../../cmd/tui generate elements.gsx
//	go run .
package main

import (
	"fmt"
	"os"

	tui "github.com/grindlemire/go-tui"
	"github.com/grindlemire/go-tui/internal/debug"
)

//go:generate go run ../../cmd/tui generate elements.gsx

func main() {
	// Initialize debug logging for focus debugging
	if err := debug.Init("debug.log"); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to init debug: %v\n", err)
	}
	defer debug.Close()
	debug.Log("=== 05-elements starting ===")

	app, err := tui.NewApp(
		tui.WithRootComponent(Elements()),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create app: %v\n", err)
		os.Exit(1)
	}
	defer app.Close()

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "App error: %v\n", err)
		os.Exit(1)
	}
}
