// Package main demonstrates a Claude Code-like chat interface using GSX.
// Type messages in the input box, press Enter to send to Claude CLI.
// Response streams above the widget. Press Escape to quit.
//
// Run `go generate` to regenerate chat_gsx.go from chat.gsx.
//
// To build and run:
//
//	cd examples/claude-chat
//	go run ../../cmd/tui generate chat.gsx
//	go run .
package main

import (
	"fmt"
	"os"

	"github.com/grindlemire/go-tui/pkg/tui"
)

//go:generate go run ../../cmd/tui generate chat.gsx

func main() {
	buf := NewTextBuffer()
	var app *tui.App

	// Update view callback - called when buffer changes
	updateView := func() {
		width, _ := app.Size()
		lines := buf.GetDisplayLines(width)
		view := ChatInput(lines)
		app.SetRoot(view.Root)
	}

	var err error
	app, err = tui.NewApp(
		tui.WithInlineHeight(5),
		tui.WithCursor(),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer app.Close()

	// Set key handler with access to app and updateView
	app.SetGlobalKeyHandler(CreateKeyHandler(buf, app, updateView))

	// Set initial view
	updateView()

	err = app.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "App error: %v\n", err)
		os.Exit(1)
	}
}
