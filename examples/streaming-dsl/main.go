// Package main demonstrates streaming text using DSL-generated components
// with the new event handling system (watchers, Run loop, SetRoot).
//
// This example shows:
// - Using the Run() event loop
// - App-level global key handler for quit
// - Using SetRoot with Viewable interface
// - Named element refs for accessing elements
//
// To build and run:
//
//	cd examples/streaming-dsl
//	go run ../../cmd/tui generate streaming.tui
//	go run .
package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/grindlemire/go-tui/pkg/layout"
	"github.com/grindlemire/go-tui/pkg/tui"
	"github.com/grindlemire/go-tui/pkg/tui/element"
)

//go:generate go run ../../cmd/tui generate streaming.tui

// lineCount and elapsed are updated by channels
var lineCount int
var elapsed int

func main() {
	app, err := tui.NewApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create app: %v\n", err)
		os.Exit(1)
	}
	defer app.Close()

	// Create the streaming content element manually
	// (we can't use DSL watchers fully yet since State isn't implemented)
	content := element.New(
		element.WithDirection(layout.Column),
		element.WithBorderStyle(tui.NewStyle().Foreground(tui.Cyan)),
		element.WithBorder(tui.BorderSingle),
		element.WithFlexGrow(1),
		element.WithScrollable(element.ScrollVertical),
		element.WithFocusable(true),
	)

	// Build the UI
	view := buildUI(content)
	app.SetRoot(view)

	// Set up app-level key handler for quit and scroll
	app.SetGlobalKeyHandler(func(e tui.KeyEvent) bool {
		if e.Rune == 'q' || e.Key == tui.KeyEscape {
			app.Stop()
			return true // Event consumed
		}
		// Handle scroll keys
		if e.Rune == 'j' || e.Rune == 106 {
			content.ScrollBy(0, 1)
			return true
		}
		if e.Rune == 'k' || e.Rune == 107 {
			content.ScrollBy(0, -1)
			return true
		}
		return false // Pass to focused element
	})

	// Create data channel for streaming
	dataCh := make(chan string, 100)

	// Set up channel watcher manually (since we don't have DSL support for State yet)
	channelWatcher := tui.Watch(dataCh, func(line string) {
		lineCount++
		lineElem := element.New(
			element.WithText(line),
			element.WithTextStyle(tui.NewStyle().Foreground(tui.Green)),
		)
		content.AddChild(lineElem)
		content.ScrollToBottom()
	})
	channelWatcher.Start(app.EventQueue(), app.StopCh())

	// Set up timer watcher manually
	timerWatcher := tui.OnTimer(time.Second, func() {
		elapsed++
		tui.MarkDirty() // Force re-render to update footer
	})
	timerWatcher.Start(app.EventQueue(), app.StopCh())

	// Start the simulated streaming process
	go produce(dataCh)

	// Run the main event loop
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "App error: %v\n", err)
		os.Exit(1)
	}
}

// buildUI creates the full UI tree
func buildUI(content *element.Element) *element.Element {
	// Header
	header := element.New(
		element.WithBorderStyle(tui.NewStyle().Foreground(tui.Blue)),
		element.WithBorder(tui.BorderSingle),
		element.WithHeight(3),
		element.WithDirection(layout.Row),
		element.WithJustify(layout.JustifyCenter),
		element.WithAlign(layout.AlignCenter),
	)
	headerText := element.New(
		element.WithText("Streaming DSL Demo - Use j/k to scroll, q to quit"),
		element.WithTextStyle(tui.NewStyle().Bold().Foreground(tui.White)),
	)
	header.AddChild(headerText)

	// Footer (will be updated dynamically)
	footer := element.New(
		element.WithBorderStyle(tui.NewStyle().Foreground(tui.Blue)),
		element.WithBorder(tui.BorderSingle),
		element.WithHeight(3),
		element.WithDirection(layout.Row),
		element.WithJustify(layout.JustifyCenter),
		element.WithAlign(layout.AlignCenter),
	)
	footer.SetOnUpdate(func() {
		// Update footer text each render
		footer.RemoveAllChildren()
		footerText := element.New(
			element.WithText(fmt.Sprintf("Lines: %d | Elapsed: %ds | Press q to exit", lineCount, elapsed)),
			element.WithTextStyle(tui.NewStyle().Foreground(tui.White)),
		)
		footer.AddChild(footerText)
	})

	// Root container
	root := element.New(
		element.WithDirection(layout.Column),
	)
	root.AddChild(header, content, footer)

	return root
}

// produce sends timestamped log lines to the channel.
func produce(ch chan<- string) {
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
		ch <- fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), msg)
		time.Sleep(300 * time.Millisecond)
	}

	// Simulate ongoing activity
	requestNum := 1
	for i := 0; i < 100; i++ {
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
		ch <- line
		time.Sleep(200 * time.Millisecond)
	}

	ch <- fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), strings.Repeat("=", 40))
	ch <- fmt.Sprintf("[%s] Process completed successfully!", time.Now().Format("15:04:05"))
}
