package main

import (
	"fmt"
	"os"

	tui "github.com/grindlemire/go-tui"
)

func main() {
	app, err := tui.NewApp(
		tui.WithRootComponent(Counter()),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	defer app.Close()
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
