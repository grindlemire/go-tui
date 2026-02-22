// Package main demonstrates the Scrolling guide example.
//
// To build and run:
//
//	go run ../../../cmd/tui generate scrolling.gsx
//	go run .
package main

import (
	"fmt"
	"os"

	tui "github.com/grindlemire/go-tui"
)

//go:generate go run ../../../cmd/tui generate scrolling.gsx

func main() {
	app, err := tui.NewApp(
		tui.WithRootComponent(FileList()),
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
