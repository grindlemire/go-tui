// Package main demonstrates hover callbacks (onMouseEnter/onMouseLeave).
//
// To build and run:
//
//	go run ../../cmd/tui generate hover.gsx
//	go run .
package main

import (
	"fmt"
	"os"

	tui "github.com/grindlemire/go-tui"
)

//go:generate go run ../../cmd/tui generate hover.gsx

func main() {
	app, err := tui.NewApp(
		tui.WithMouseMotion(),
		tui.WithRootComponent(HoverDemo()),
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
