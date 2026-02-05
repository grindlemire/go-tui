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

	app, err := tui.NewApp(
		tui.WithRoot(Loops(items)),
		tui.WithGlobalKeyHandler(func(e tui.KeyEvent) bool {
			if e.Rune == 'q' || e.Key == tui.KeyEscape {
				tui.Stop()
				return true
			}
			return false
		}),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create app: %v\n", err)
		os.Exit(1)
	}
	defer app.Close()

	err = app.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "App error: %v\n", err)
		os.Exit(1)
	}
}
