// main.go - Minimal bootstrap for the streaming counter example
//
// With Refs, State, and Event Handling, main.go is just setup and app.Run().
// All behavior is declared in the .tui file.

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/grindlemire/go-tui/pkg/tui"
)

//go:generate go run ../../cmd/tui generate state.tui

func main() {
	app, err := tui.NewApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create app: %v\n", err)
		os.Exit(1)
	}
	defer app.Close()

	// Create data channel for streaming content
	dataCh := make(chan string, 100)
	go produceData(dataCh)

	counter := StreamingCounter(dataCh)

	// Set the root component and set the global key listener
	app.SetRoot(counter)
	app.OnKeyEvent(func(e tui.KeyEvent) {
		if e.Rune == 'q' {
			app.Stop()
			return
		}
	})

	// Run blocks until tui.Stop() is called
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "App error: %v\n", err)
		os.Exit(1)
	}
}

func produceData(ch chan<- string) {
	defer close(ch)
	for i := 0; i < 100; i++ {
		ch <- fmt.Sprintf("[%s] Log line %d - some sample streaming data",
			time.Now().Format("15:04:05"), i)
		time.Sleep(200 * time.Millisecond)
	}
}
