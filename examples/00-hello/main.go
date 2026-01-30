// Package main demonstrates the most basic go-tui example.
//
// To build and run:
//
//	go run ../../cmd/tui generate hello.gsx
//	go run .
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/grindlemire/go-tui/pkg/layout"
	"github.com/grindlemire/go-tui/pkg/tui"
	"github.com/grindlemire/go-tui/pkg/tui/element"
)

//go:generate go run ../../cmd/tui generate hello.gsx

func main() {
	app, err := tui.NewApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create app: %v\n", err)
		os.Exit(1)
	}
	defer app.Close()

	root := buildUI(app)
	app.SetRoot(root)

	for {
		event, ok := app.PollEvent(50 * time.Millisecond)
		if ok {
			switch e := event.(type) {
			case tui.KeyEvent:
				if e.Key == tui.KeyEscape || e.Rune == 'q' {
					return
				}
			case tui.ResizeEvent:
				root = buildUI(app)
				app.SetRoot(root)
			}
		}
		app.Render()
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

	hello := Hello()
	root.AddChild(hello.Root)

	return root
}
