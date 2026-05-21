// Package main is a diagnostic example that prints every event the app
// receives. Useful for validating key, mouse, and resize handling on a new
// platform (notably Windows after the input-record rewrite).
//
// To build and run:
//
//	go run ../../cmd/tui generate dump.gsx
//	go run .
package main

import (
	"fmt"
	"os"

	tui "github.com/grindlemire/go-tui"
)

//go:generate go run ../../cmd/tui generate dump.gsx

func main() {
	app, err := tui.NewApp(
		tui.WithRootComponent(EventDump()),
		tui.WithMouse(),
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
