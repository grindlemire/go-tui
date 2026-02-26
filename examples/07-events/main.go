// Package main demonstrates the Event Handling guide example.
//
// To build and run:
//
//	go run ../../../cmd/tui generate events.gsx
//	go run .
package main

import (
	"fmt"
	"os"

	tui "github.com/grindlemire/go-tui"
)

//go:generate go run ../../../cmd/tui generate events.gsx

func main() {
	app, err := tui.NewApp(
		tui.WithRootComponent(Explorer()),
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
