// Package main demonstrates text styling with colors and fonts.
//
// To build and run:
//
//	go run ../../cmd/tui generate styling.gsx
//	go run .
package main

import (
	"fmt"
	"os"

	tui "github.com/grindlemire/go-tui"
)

//go:generate go run ../../cmd/tui generate styling.gsx

func main() {
	app, err := tui.NewApp(
		tui.WithRoot(Styling()),
		tui.WithGlobalKeyHandler(func(e tui.KeyEvent) bool {
			if e.Key == tui.KeyEscape || e.Rune == 'q' {
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
	fmt.Println("App exited successfully!")
}
