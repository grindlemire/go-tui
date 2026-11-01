// Package main demonstrates hover: hover: classes, onMouseEnter/onMouseLeave
// callbacks, keyed lists, component tags, drags, and edge cases.
//
// To build and run:
//
//	go run ../../cmd/tui generate .
//	go run .          # full screen
//	go run . -inline  # inline mode (24 rows below the prompt)
package main

import (
	"flag"
	"fmt"
	"os"

	tui "github.com/grindlemire/go-tui"
)

//go:generate go run ../../cmd/tui generate .

func main() {
	inline := flag.Bool("inline", false, "run in inline mode instead of full screen")
	flag.Parse()

	opts := []tui.AppOption{
		tui.WithRootComponent(HoverDemo(*inline)),
	}
	if *inline {
		// Inline mode leaves the mouse off by default.
		opts = append(opts, tui.WithInlineHeight(24), tui.WithMouse())
	}
	app, err := tui.NewApp(opts...)
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
