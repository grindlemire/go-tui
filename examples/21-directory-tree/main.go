// Package main demonstrates a foldable directory tree with keyboard navigation.
//
// To build and run:
//
//	go run ../../cmd/tui generate tree.gsx
//	go run .
package main

import (
	"fmt"
	"os"

	tui "github.com/grindlemire/go-tui"
)

//go:generate go run ../../cmd/tui generate tree.gsx

func main() {
	app, err := tui.NewApp(
		tui.WithRootComponent(DirectoryTree()),
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
