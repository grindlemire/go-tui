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
	"time"

	tui "github.com/grindlemire/go-tui"
)

//go:generate go run ../../cmd/tui generate loops.gsx

func main() {
	app, err := tui.NewApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create app: %v\n", err)
		os.Exit(1)
	}
	defer app.Close()

	items := []string{"Apple", "Banana", "Cherry", "Date", "Elderberry"}
	selected := 0

	root := buildUI(app, items, selected)
	app.SetRoot(root)

	for {
		event, ok := app.PollEvent(50 * time.Millisecond)
		if ok {
			switch e := event.(type) {
			case tui.KeyEvent:
				switch {
				case e.Key == tui.KeyEscape || e.Rune == 'q':
					return
				case e.Rune == 'j' || e.Key == tui.KeyDown:
					if selected < len(items)-1 {
						selected++
						root = buildUI(app, items, selected)
						app.SetRoot(root)
					}
				case e.Rune == 'k' || e.Key == tui.KeyUp:
					if selected > 0 {
						selected--
						root = buildUI(app, items, selected)
						app.SetRoot(root)
					}
				}
			case tui.ResizeEvent:
				root = buildUI(app, items, selected)
				app.SetRoot(root)
			}
		}
		app.Render()
	}
}

func buildUI(app *tui.App, items []string, selected int) *tui.Element {
	width, height := app.Size()

	root := tui.New(
		tui.WithSize(width, height),
		tui.WithDirection(tui.Column),
		tui.WithJustify(tui.JustifyCenter),
		tui.WithAlign(tui.AlignCenter),
	)

	loops := Loops(items, selected)
	root.AddChild(loops.Root)

	return root
}
