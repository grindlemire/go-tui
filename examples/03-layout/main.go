// Package main demonstrates flexbox layout with direction, justify, align, gap, and sizing.
//
// To build and run:
//
//	go run ../../cmd/tui generate layout.gsx
//	go run .
package main

import (
	"fmt"
	"os"

	tui "github.com/grindlemire/go-tui"
)

//go:generate go run ../../cmd/tui generate layout.gsx

func main() {
	app, err := tui.NewApp(
		tui.WithRootComponent(Layout()),
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
