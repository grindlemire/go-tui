// Package main demonstrates streaming text using DSL-generated components.
// This example shows how to combine .tui file components with imperative
// streaming logic for real-time content with auto-scroll behavior.
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

// StreamBox wraps an Element to provide channel-based text streaming.
type StreamBox struct {
	elem       *element.Element
	textCh     <-chan string
	textStyle  tui.Style
	autoScroll bool
}

// NewStreamBox creates a new StreamBox that receives text from the given channel.
func NewStreamBox(textCh <-chan string) *StreamBox {
	s := &StreamBox{
		elem: element.New(
			element.WithScrollable(element.ScrollVertical),
			element.WithDirection(layout.Column),
		),
		textCh:     textCh,
		textStyle:  tui.NewStyle().Foreground(tui.Green),
		autoScroll: true,
	}

	// Set up automatic polling via onUpdate hook
	s.elem.SetOnUpdate(s.poll)

	return s
}

// Element returns the underlying Element for adding to the tree.
func (s *StreamBox) Element() *element.Element {
	return s.elem
}

// IsAutoScrolling returns whether auto-scroll is currently enabled.
func (s *StreamBox) IsAutoScrolling() bool {
	return s.autoScroll
}

// poll is called automatically before each render via the onUpdate hook.
func (s *StreamBox) poll() {
	hadContent := false
	for {
		select {
		case text, ok := <-s.textCh:
			if !ok {
				return
			}
			s.appendText(text)
			hadContent = true
		default:
			if hadContent && s.autoScroll {
				s.elem.ScrollToBottom()
			}
			return
		}
	}
}

// appendText splits text on newlines and adds each line as a child element.
func (s *StreamBox) appendText(text string) {
	s.updateAutoScroll()

	lines := strings.Split(text, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		lineElem := element.New(
			element.WithText(line),
			element.WithTextStyle(s.textStyle),
		)
		s.elem.AddChild(lineElem)
	}
}

// updateAutoScroll checks if the user has scrolled away from the bottom.
func (s *StreamBox) updateAutoScroll() {
	_, scrollY := s.elem.ScrollOffset()
	_, maxY := s.elem.MaxScroll()

	if scrollY >= maxY-1 {
		s.autoScroll = true
	}
}

// HandleScrollEvent updates the auto-scroll state based on user interaction.
func (s *StreamBox) HandleScrollEvent() {
	_, scrollY := s.elem.ScrollOffset()
	_, maxY := s.elem.MaxScroll()

	if scrollY < maxY-1 {
		s.autoScroll = false
	} else {
		s.autoScroll = true
	}
}

func main() {
	app, err := tui.NewApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create app: %v\n", err)
		os.Exit(1)
	}
	defer app.Close()

	width, height := app.Size()

	// Create a channel for streaming text
	textCh := make(chan string, 100)

	// Create the StreamBox for content
	streamBox := NewStreamBox(textCh)
	streamBox.Element().SetBorder(tui.BorderSingle)
	streamBox.Element().SetBorderStyle(tui.NewStyle().Foreground(tui.Cyan))

	// Apply flexGrow to make it fill available space
	style := streamBox.Element().Style()
	style.FlexGrow = 1
	style.Padding = layout.EdgeAll(1)
	streamBox.Element().SetStyle(style)

	// Build UI using DSL-generated components
	view := buildUI(width, height, streamBox.Element(), 0, 0, "ON")
	app.SetRoot(view.Root)

	// Register streamBox for focus (so it receives scroll events)
	app.Focus().Register(streamBox.Element())

	// Start the simulated streaming process
	go simulateProcess(textCh)

	// Main event loop
	for {
		event, ok := app.PollEvent(50 * time.Millisecond)
		if ok {
			switch e := event.(type) {
			case tui.KeyEvent:
				if e.Key == tui.KeyEscape {
					return
				}
				// Handle vim-style navigation
				if e.Key == tui.KeyRune {
					switch e.Rune {
					case 'j':
						streamBox.Element().ScrollBy(0, 1)
						streamBox.HandleScrollEvent()
					case 'k':
						streamBox.Element().ScrollBy(0, -1)
						streamBox.HandleScrollEvent()
					case 'g':
						streamBox.Element().ScrollToTop()
						streamBox.HandleScrollEvent()
					case 'G':
						streamBox.Element().ScrollToBottom()
						streamBox.HandleScrollEvent()
					}
				}
				// Track scrolling for auto-scroll behavior
				switch e.Key {
				case tui.KeyUp, tui.KeyDown, tui.KeyPageUp, tui.KeyPageDown, tui.KeyHome, tui.KeyEnd:
					app.Dispatch(event)
					streamBox.HandleScrollEvent()
				default:
					app.Dispatch(event)
				}

			case tui.ResizeEvent:
				width, height = e.Width, e.Height
				app.Dispatch(event)
			}
		}

		// Get scroll metrics for footer
		_, scrollY := streamBox.Element().ScrollOffset()
		_, contentH := streamBox.Element().ContentSize()
		_, viewportH := streamBox.Element().ViewportSize()
		maxScroll := max(0, contentH-viewportH)
		autoScrollStatus := "OFF"
		if streamBox.IsAutoScrolling() {
			autoScrollStatus = "ON"
		}

		// Rebuild UI with updated footer (DSL components are cheap to rebuild)
		view = buildUI(width, height, streamBox.Element(), scrollY, maxScroll, autoScrollStatus)
		app.SetRoot(view.Root)

		app.Render()
	}
}

// UIView is the view struct for the manually created UI
type UIView struct {
	Root *element.Element
}

// buildUI creates the UI tree using DSL-generated Header and Footer components.
func buildUI(width, height int, content *element.Element, scrollY, maxScroll int, autoScrollStatus string) UIView {
	root := element.New(
		element.WithSize(width, height),
		element.WithDirection(layout.Column),
	)

	// Use DSL-generated components for header and footer - now return view structs
	header := Header()
	footer := Footer(scrollY, maxScroll, autoScrollStatus)

	root.AddChild(header.Root, content, footer.Root)

	return UIView{Root: root}
}

// simulateProcess sends timestamped log lines to the channel.
func simulateProcess(ch chan<- string) {
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
		switch i % 5 {
		case 0:
			ch <- fmt.Sprintf("[%s] Processing request #%d", time.Now().Format("15:04:05"), requestNum)
			requestNum++
		case 1:
			ch <- fmt.Sprintf("[%s] Database query completed in %dms", time.Now().Format("15:04:05"), 10+i%50)
		case 2:
			ch <- fmt.Sprintf("[%s] Cache hit ratio: %.1f%%", time.Now().Format("15:04:05"), 85.0+float64(i%10))
		case 3:
			ch <- fmt.Sprintf("[%s] Memory usage: %dMB", time.Now().Format("15:04:05"), 128+i*2)
		case 4:
			ch <- fmt.Sprintf("[%s] Active connections: %d", time.Now().Format("15:04:05"), 50+i%20)
		}
		time.Sleep(200 * time.Millisecond)
	}

	ch <- fmt.Sprintf("[%s] Process completed successfully!", time.Now().Format("15:04:05"))
}
