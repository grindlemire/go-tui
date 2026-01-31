// Package main demonstrates component composition.
//
// To build and run:
//
//	go run ../../cmd/tui generate composition.gsx
//	go run .
package main

import (
	"fmt"
	"os"
	"time"

	tui "github.com/grindlemire/go-tui"
)

//go:generate go run ../../cmd/tui generate composition.gsx

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

func buildUI(app *tui.App) *tui.Element {
	width, height := app.Size()

	root := tui.New(
		tui.WithSize(width, height),
		tui.WithDirection(tui.Column),
	)

	appView := App()
	root.AddChild(appView.Root)

	return root
}
