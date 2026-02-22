// Package main demonstrates the Building a Dashboard guide example.
//
// To build and run:
//
//	go run ../../../cmd/tui generate dashboard.gsx
//	go run .
package main

import (
	"fmt"
	"os"

	tui "github.com/grindlemire/go-tui"
)

//go:generate go run ../../../cmd/tui generate dashboard.gsx

func main() {
	eventCh := make(chan string, 100)

	app, err := tui.NewApp(
		tui.WithRootComponent(Dashboard(eventCh)),
		tui.WithMouse(),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create app: %v\n", err)
		os.Exit(1)
	}
	defer app.Close()

	go produceEvents(eventCh, app.StopCh())

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "App error: %v\n", err)
		os.Exit(1)
	}
}
