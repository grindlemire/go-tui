package main

import (
	"fmt"
	"os"

	tui "github.com/grindlemire/go-tui"
	"golang.org/x/term"
)

func main() {
	// Get terminal width for text wrapping
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width = 80 // fallback
	}

	app, err := tui.NewApp(
		tui.WithInlineHeight(3), // Start with 3, will grow as needed
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer app.Close()

	app.SetRoot(Chat(width))

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
