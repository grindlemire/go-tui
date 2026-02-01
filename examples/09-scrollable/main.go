// Package main demonstrates scrollable content.
//
// To build and run:
//
//	go run ../../cmd/tui generate scrollable.gsx
//	go run .
package main

import (
	"fmt"
	"os"

	tui "github.com/grindlemire/go-tui"
)

//go:generate go run ../../cmd/tui generate scrollable.gsx

func main() {
	// Generate sample items
	items := make([]string, 30)
	for i := range items {
		items[i] = fmt.Sprintf("Sample item with some content here")
	}

	view := Scrollable(items)

	app, err := tui.NewApp(
		tui.WithRoot(view),
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

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "App error: %v\n", err)
		os.Exit(1)
	}
}

func itemStyle(i int) string {
	if i%2 == 0 {
		return "text-green"
	}
	return "text-yellow"
}
