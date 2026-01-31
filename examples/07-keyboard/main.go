// Package main demonstrates keyboard event handling with onKeyPress.
//
// To build and run:
//
//	go run ../../cmd/tui generate keyboard.gsx
//	go run .
package main

import (
	"fmt"
	"os"

	tui "github.com/grindlemire/go-tui"
)

//go:generate go run ../../cmd/tui generate keyboard.gsx

func main() {
	app, err := tui.NewApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create app: %v\n", err)
		os.Exit(1)
	}
	defer app.Close()

	root := buildUI(app)
	app.SetRoot(root)

	app.SetGlobalKeyHandler(func(e tui.KeyEvent) bool {
		if e.Rune == 'q' || e.Key == tui.KeyEscape {
			app.Stop()
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

func buildUI(app *tui.App) *tui.Element {
	width, height := app.Size()

	root := tui.New(
		tui.WithSize(width, height),
		tui.WithDirection(tui.Column),
		tui.WithJustify(tui.JustifyCenter),
		tui.WithAlign(tui.AlignCenter),
	)

	keyboard := Keyboard()
	root.AddChild(keyboard.Root)

	return root
}
