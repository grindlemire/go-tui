// Package main demonstrates streaming text using DSL-generated components
// with reactive state and explicit refs.
//
// This example shows:
// - Fully declarative UI in .gsx file
// - Reactive state with auto-binding for footer updates
// - Explicit refs (ref={content}) for cross-element access in handlers
// - Self-inject handlers for element-local operations
// - onChannel and onTimer watchers in DSL
//
// To build and run:
//
//	cd examples/streaming-dsl
//	go run ../../cmd/tui generate streaming.gsx
//	go run .
package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	tui "github.com/grindlemire/go-tui"
)

//go:generate go run ../../cmd/tui generate streaming.gsx

func main() {
	// Create data channel for streaming
	dataCh := make(chan string, 100)

	// Build fully declarative UI - all handlers defined in .gsx
	view := StreamApp(dataCh)

	// Create app with all configuration via options
	app, err := tui.NewApp(
		tui.WithRootView(view),
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

	// Start the simulated streaming process
	go produce(dataCh, app.StopCh())

	// Run the main event loop
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "App error: %v\n", err)
		os.Exit(1)
	}
}

// produce sends timestamped log lines to the channel.
func produce(ch chan<- string, stopCh <-chan struct{}) {
	defer close(ch)

	// Initial startup messages
	messages := []string{
		"Process started...",
		"Initializing components...",
		"Loading configuration...",
		"Connecting to services...",
		"Ready to process requests",
	}

	for _, msg := range messages {
		select {
		case <-stopCh:
			return
		case ch <- fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), msg):
		}
		time.Sleep(300 * time.Millisecond)
	}

	// Simulate ongoing activity
	requestNum := 1
	for i := 0; i < 100; i++ {
		select {
		case <-stopCh:
			return
		default:
		}

		var line string
		switch i % 5 {
		case 0:
			line = fmt.Sprintf("[%s] Processing request #%d", time.Now().Format("15:04:05"), requestNum)
			requestNum++
		case 1:
			line = fmt.Sprintf("[%s] Database query completed in %dms", time.Now().Format("15:04:05"), 10+i%50)
		case 2:
			line = fmt.Sprintf("[%s] Cache hit ratio: %.1f%%", time.Now().Format("15:04:05"), 85.0+float64(i%10))
		case 3:
			line = fmt.Sprintf("[%s] Memory usage: %dMB", time.Now().Format("15:04:05"), 128+i*2)
		case 4:
			line = fmt.Sprintf("[%s] Active connections: %d", time.Now().Format("15:04:05"), 50+i%20)
		}

		select {
		case <-stopCh:
			return
		case ch <- line:
		}
		time.Sleep(200 * time.Millisecond)
	}

	select {
	case <-stopCh:
		return
	case ch <- fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), strings.Repeat("=", 40)):
	}
	select {
	case <-stopCh:
		return
	case ch <- fmt.Sprintf("[%s] Process completed successfully!", time.Now().Format("15:04:05")):
	}
}
