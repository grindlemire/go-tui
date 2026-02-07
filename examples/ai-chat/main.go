package main

import (
	"fmt"
	"os"

	tui "github.com/grindlemire/go-tui"
)

func main() {
	app, err := tui.NewApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create app: %v\n", err)
		os.Exit(1)
	}
	defer app.Close()

	// Placeholder root
	root := tui.New(
		tui.WithText("AI Chat - Press q to quit"),
	)
	app.SetRoot(root)

	// Simple exit on 'q'
	app.SetGlobalKeyHandler(func(ke tui.KeyEvent) bool {
		if ke.Rune == 'q' {
			app.Stop()
			return true
		}
		return false
	})

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "App error: %v\n", err)
		os.Exit(1)
	}
}
