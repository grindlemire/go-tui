// Package main tests inline mode behavior.
// Press 'p' to print a line above, 'q' or Escape to quit.
package main

import (
	"fmt"
	"os"

	tui "github.com/grindlemire/go-tui"
)

//go:generate go run ../../cmd/tui generate widget.gsx

const inlineHeight = 3

func main() {
	tuiApp, err := tui.NewApp(
		tui.WithInlineHeight(inlineHeight),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer tuiApp.Close()

	tuiApp.SetRootComponent(InlineApp(tuiApp, inlineHeight))

	if err := tuiApp.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
