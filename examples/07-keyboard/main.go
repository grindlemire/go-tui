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

	"github.com/grindlemire/go-tui/pkg/layout"
	"github.com/grindlemire/go-tui/pkg/tui"
	"github.com/grindlemire/go-tui/pkg/tui/element"
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

func buildUI(app *tui.App) *element.Element {
	width, height := app.Size()

	root := element.New(
		element.WithSize(width, height),
		element.WithDirection(layout.Column),
		element.WithJustify(layout.JustifyCenter),
		element.WithAlign(layout.AlignCenter),
	)

	keyboard := Keyboard()
	root.AddChild(keyboard.Root)

	return root
}
