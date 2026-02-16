// Package main demonstrates component composition with templ, children, and reuse.
//
// To build and run:
//
//	go run ../../cmd/tui generate components.gsx
//	go run .
package main

import (
	"fmt"
	"os"

	tui "github.com/grindlemire/go-tui"
)

//go:generate go run ../../cmd/tui generate components.gsx

func main() {
	app, err := tui.NewApp(
		tui.WithRootComponent(App()),
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
