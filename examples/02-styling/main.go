// Package main demonstrates text styling with colors, fonts, borders, and gradients.
//
// To build and run:
//
//	go run ../../cmd/tui generate styling.gsx
//	go run .
package main

import (
	"fmt"
	"os"

	tui "github.com/grindlemire/go-tui"
)

//go:generate go run ../../cmd/tui generate styling.gsx

func main() {
	app, err := tui.NewApp(
		tui.WithRootComponent(Styling()),
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
