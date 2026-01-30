// Package main demonstrates named element references (#Name).
//
// Named refs allow imperative access to elements from handlers.
// This example shows how to use refs to access elements by name.
//
// To build and run:
//
//	go run ../../cmd/tui generate refs.gsx
//	go run .
package main

import (
	"fmt"
	"os"

	"github.com/grindlemire/go-tui/pkg/layout"
	"github.com/grindlemire/go-tui/pkg/tui"
	"github.com/grindlemire/go-tui/pkg/tui/element"
)

//go:generate go run ../../cmd/tui generate refs.gsx

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

	// The returned view struct contains named refs as fields:
	// refs.Counter, refs.IncrementBtn, refs.DecrementBtn, refs.Status
	refs := Refs()
	root.AddChild(refs.Root)

	// You can access named refs for imperative operations:
	// refs.Counter.SetBorder(tui.BorderDouble)
	// refs.Status.SetText("Updated status")

	return root
}
