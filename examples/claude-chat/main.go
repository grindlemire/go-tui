// Package main demonstrates InlineHeight mode with a simple input box.
// Text streams above the widget while you can type in the input box.
// Press Enter to add your input to the stream. Press Escape to quit.
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
	"time"

	tui "github.com/grindlemire/go-tui"
)

//go:generate go run ../../cmd/tui generate .

const inlineHeight = 5

func main() {
	var input string

	// Create app with inline height (widget at bottom of terminal)
	app, err := tui.NewApp(
		tui.WithInlineHeight(inlineHeight),
		tui.WithCursor(),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer app.Close()

	// Start streaming random text above the widget
	textCh := make(chan string, 100)
	go generateText(textCh)
	go func() {
		for text := range textCh {
			app.QueuePrintAboveln("%s", text)
		}
	}()

	// Render function updates the UI
	render := func() {
		view := InputBox(input, inlineHeight)
		app.SetRoot(view.Root)
	}

	// Handle keyboard input
	app.SetGlobalKeyHandler(func(e tui.KeyEvent) bool {
		switch e.Key {
		case tui.KeyRune:
			input += string(e.Rune)
		case tui.KeyBackspace:
			if len(input) > 0 {
				input = input[:len(input)-1]
			}
		case tui.KeyEnter:
			if input != "" {
				app.PrintAboveln("You: %s", input)
				input = ""
			}
		case tui.KeyEscape:
			app.Stop()
			return true
		default:
			return false
		}
		render()
		tui.MarkDirty()
		return true
	})

	render()
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// generateText sends timestamped messages to the channel.
// Replace this with any text source (API responses, logs, etc).
func generateText(ch chan<- string) {
	words := []string{"alpha", "beta", "gamma", "delta", "epsilon"}
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	count := 1
	for range ticker.C {
		word := words[count%len(words)]
		ch <- fmt.Sprintf("[%s] Message #%d: %s",
			time.Now().Format("15:04:05"), count, word)
		count++
	}
}
