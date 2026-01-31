// Package main demonstrates conditional rendering with @if/@else.
//
// To build and run:
//
//	go run ../../cmd/tui generate conditionals.gsx
//	go run .
package main

import (
	"fmt"
	"os"
	"time"

	tui "github.com/grindlemire/go-tui"
)

//go:generate go run ../../cmd/tui generate conditionals.gsx

func main() {
	app, err := tui.NewApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create app: %v\n", err)
		os.Exit(1)
	}
	defer app.Close()

	enabled := true
	root := buildUI(app, enabled)
	app.SetRoot(root)

	for {
		event, ok := app.PollEvent(50 * time.Millisecond)
		if ok {
			switch e := event.(type) {
			case tui.KeyEvent:
				switch {
				case e.Key == tui.KeyEscape || e.Rune == 'q':
					return
				case e.Rune == 't':
					enabled = !enabled
					root = buildUI(app, enabled)
					app.SetRoot(root)
				}
			case tui.ResizeEvent:
				root = buildUI(app, enabled)
				app.SetRoot(root)
			}
		}
		app.Render()
	}
}

func buildUI(app *tui.App, enabled bool) *tui.Element {
	width, height := app.Size()

	root := tui.New(
		tui.WithSize(width, height),
		tui.WithDirection(tui.Column),
		tui.WithJustify(tui.JustifyCenter),
		tui.WithAlign(tui.AlignCenter),
	)

	conditionals := Conditionals(enabled)
	root.AddChild(conditionals.Root)

	return root
}
