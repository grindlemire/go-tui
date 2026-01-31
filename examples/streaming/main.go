// Package main demonstrates streaming text into a scrollable TUI element.
// This example shows how to use the onUpdate hook to poll a channel
// and display real-time streaming content with auto-scroll behavior.
package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	tui "github.com/grindlemire/go-tui"
)

// StreamBox wraps an Element to provide channel-based text streaming.
// This is defined locally in the example - not part of the library.
// Users can copy and adapt this pattern for their own streaming needs.
type StreamBox struct {
	elem       *tui.Element
	textCh     <-chan string
	textStyle  tui.Style
	autoScroll bool
}

// NewStreamBox creates a new StreamBox that receives text from the given channel.
// The StreamBox automatically polls the channel before each render and
// displays each line as a child element with auto-scroll behavior.
func NewStreamBox(textCh <-chan string) *StreamBox {
	s := &StreamBox{
		elem: tui.New(
			tui.WithScrollable(tui.ScrollVertical),
			tui.WithDirection(tui.Column),
		),
		textCh:     textCh,
		textStyle:  tui.NewStyle().Foreground(tui.White),
		autoScroll: true,
	}

	// Set up automatic polling via onUpdate hook
	s.elem.SetOnUpdate(s.poll)

	return s
}

// Element returns the underlying Element for adding to the tree.
func (s *StreamBox) Element() *tui.Element {
	return s.elem
}

// SetTextStyle sets the style used for new text lines.
func (s *StreamBox) SetTextStyle(style tui.Style) {
	s.textStyle = style
}

// IsAutoScrolling returns whether auto-scroll is currently enabled.
func (s *StreamBox) IsAutoScrolling() bool {
	return s.autoScroll
}

// poll is called automatically before each render via the onUpdate hook.
// It drains all available messages from the channel.
func (s *StreamBox) poll() {
	hadContent := false
	for {
		select {
		case text, ok := <-s.textCh:
			if !ok {
				// Channel closed - nothing more to receive
				return
			}
			s.appendText(text)
			hadContent = true
		default:
			// No more messages available
			if hadContent && s.autoScroll {
				s.elem.ScrollToBottom()
			}
			return
		}
	}
}

// appendText splits text on newlines and adds each line as a child element.
func (s *StreamBox) appendText(text string) {
	// Check if we're at the bottom before adding content
	// This determines whether we should auto-scroll
	s.updateAutoScroll()

	lines := strings.Split(text, "\n")
	for _, line := range lines {
		// Skip empty lines at the end (from trailing newline)
		if line == "" {
			continue
		}
		lineElem := tui.New(
			tui.WithText(line),
			tui.WithTextStyle(s.textStyle),
		)
		s.elem.AddChild(lineElem)
	}
}

// updateAutoScroll checks if the user has scrolled away from the bottom.
// If they have, we disable auto-scroll. If they scroll back to bottom,
// we re-enable it.
func (s *StreamBox) updateAutoScroll() {
	_, scrollY := s.elem.ScrollOffset()
	_, maxY := s.elem.MaxScroll()

	// User is "at bottom" if within 1 line of max scroll
	atBottom := scrollY >= maxY-1

	if atBottom {
		s.autoScroll = true
	}
}

// HandleScrollEvent should be called when the StreamBox receives a scroll event.
// It updates the auto-scroll state based on user interaction.
func (s *StreamBox) HandleScrollEvent() {
	_, scrollY := s.elem.ScrollOffset()
	_, maxY := s.elem.MaxScroll()

	// If user scrolled up (not at bottom), disable auto-scroll
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

	// Root container
	root := tui.New(
		tui.WithSize(width, height),
		tui.WithDirection(tui.Column),
	)

	// Header
	header := tui.New(
		tui.WithHeight(3),
		tui.WithDirection(tui.Row),
		tui.WithJustify(tui.JustifyCenter),
		tui.WithAlign(tui.AlignCenter),
		tui.WithBorder(tui.BorderSingle),
		tui.WithBorderStyle(tui.NewStyle().Foreground(tui.Blue)),
	)
	headerTitle := tui.New(
		tui.WithText("Streaming Demo - Use j/k, PgUp/PgDn, Arrow Keys to scroll"),
		tui.WithTextStyle(tui.NewStyle().Foreground(tui.White).Bold()),
	)
	header.AddChild(headerTitle)

	// Create a channel for streaming text
	textCh := make(chan string, 100)

	// Create the StreamBox
	streamBox := NewStreamBox(textCh)
	streamBox.SetTextStyle(tui.NewStyle().Foreground(tui.Green))
	streamBox.Element().SetBorder(tui.BorderSingle)
	streamBox.Element().SetBorderStyle(tui.NewStyle().Foreground(tui.Cyan))

	// Apply flexGrow to make it fill available space
	style := streamBox.Element().Style()
	style.FlexGrow = 1
	style.Padding = tui.EdgeAll(1)
	streamBox.Element().SetStyle(style)

	// Footer with status
	footer := tui.New(
		tui.WithHeight(3),
		tui.WithDirection(tui.Row),
		tui.WithJustify(tui.JustifyCenter),
		tui.WithAlign(tui.AlignCenter),
		tui.WithBorder(tui.BorderSingle),
		tui.WithBorderStyle(tui.NewStyle().Foreground(tui.Blue)),
	)
	footerText := tui.New(
		tui.WithText("Press ESC to exit"),
		tui.WithTextStyle(tui.NewStyle().Foreground(tui.White)),
	)
	footer.AddChild(footerText)

	root.AddChild(header, streamBox.Element(), footer)
	app.SetRoot(root)

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
					// Dispatch first, then check scroll position
					app.Dispatch(event)
					streamBox.HandleScrollEvent()
				default:
					app.Dispatch(event)
				}

			case tui.ResizeEvent:
				width, height = e.Width, e.Height
				style := root.Style()
				style.Width = tui.Fixed(width)
				style.Height = tui.Fixed(height)
				root.SetStyle(style)
				app.Dispatch(event)
			}
		}

		// Update footer with scroll position and auto-scroll status
		_, y := streamBox.Element().ScrollOffset()
		_, contentH := streamBox.Element().ContentSize()
		_, viewportH := streamBox.Element().ViewportSize()
		autoScrollStatus := "OFF"
		if streamBox.IsAutoScrolling() {
			autoScrollStatus = "ON"
		}
		footerText.SetText(fmt.Sprintf("Scroll: %d/%d | Auto-scroll: %s | Press ESC to exit",
			y, max(0, contentH-viewportH), autoScrollStatus))

		app.Render()
	}
}

// simulateProcess sends timestamped log lines to the channel,
// simulating a long-running process producing output.
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
		// Vary the message types
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
