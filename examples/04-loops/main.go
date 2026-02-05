// Package main demonstrates loop rendering with @for.
//
// To build and run:
//
//	go run ../../cmd/tui generate loops.gsx
//	go run .
package main

import (
	"fmt"
	"os"

	tui "github.com/grindlemire/go-tui"
)

//go:generate go run ../../cmd/tui generate loops.gsx

func main() {
	items := []string{"Apple", "Banana", "Cherry", "Date", "Elderberry"}

	app, err := tui.NewApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create app: %v\n", err)
		os.Exit(1)
	}
	defer app.Close()

	app.SetRoot(Loops(items))

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "App error: %v\n", err)
		os.Exit(1)
	}
}
