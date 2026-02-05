// Package main demonstrates interactive elements with all event listeners.
//
// To build and run:
//
//	go run ../../cmd/tui generate interactive.gsx
//	go run .
package main

import (
	"fmt"
	"os"

	tui "github.com/grindlemire/go-tui"
)

//go:generate go run ../../cmd/tui generate interactive.gsx

func main() {
	app, err := tui.NewApp(
		tui.WithRoot(Interactive()),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create app: %v\n", err)
		os.Exit(1)
	}
	defer app.Close()

	app.SetGlobalKeyHandler(func(e tui.KeyEvent) bool {
		if e.Rune == 'q' || e.Key == tui.KeyEscape {
			tui.Stop()
			return true
		}
		if e.Key == tui.KeyTab {
			if e.Mod.Has(tui.ModShift) {
				app.FocusPrev()
			} else {
				app.FocusNext()
			}
			return true
		}
		return false
	})

	err = app.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "App error: %v\n", err)
		os.Exit(1)
	}
}
