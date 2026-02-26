// Package main demonstrates the Inline Mode guide example.
//
// To build and run:
//
//	go run ../../../cmd/tui generate inline.gsx
//	go run .
package main

import (
	"fmt"
	"os"

	tui "github.com/grindlemire/go-tui"
)

//go:generate go run ../../../cmd/tui generate inline.gsx

func main() {
	app, err := tui.NewApp(
		tui.WithInlineHeight(3),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer app.Close()

	app.SetRootComponent(InlineApp())

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
