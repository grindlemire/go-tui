// Package main demonstrates refs for hit-testing and mouse click handling.
//
// To build and run:
//
//	go run ../../cmd/tui generate clicks.gsx
//	go run .
package main

import (
	"fmt"
	"os"

	tui "github.com/grindlemire/go-tui"
)

//go:generate go run ../../cmd/tui generate clicks.gsx

func main() {
	app, err := tui.NewApp(
		tui.WithRootComponent(ColorMixer()),
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
