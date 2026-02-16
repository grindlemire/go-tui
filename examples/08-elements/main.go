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
)

//go:generate go run ../../cmd/tui generate elements.gsx

func main() {
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
