package main

import (
	"fmt"
	"os"
	"time"

	"github.com/grindlemire/go-tui/pkg/layout"
	"github.com/grindlemire/go-tui/pkg/tui"
	"github.com/grindlemire/go-tui/pkg/tui/element"
)

func main() {
	term, err := tui.NewANSITerminal(os.Stdout, os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create terminal: %v\n", err)
		os.Exit(1)
	}

	if err := term.EnterRawMode(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to enter raw mode: %v\n", err)
		os.Exit(1)
	}
	defer term.ExitRawMode()

	term.EnterAltScreen()
	defer term.ExitAltScreen()

	term.HideCursor()
	defer term.ShowCursor()

	width, height := term.Size()

	// Build layout tree once - we'll mutate it for animation
	root := element.New(
		element.WithSize(width, height),
		element.WithDirection(layout.Column),
		element.WithJustify(layout.JustifyCenter),
		element.WithAlign(layout.AlignCenter),
	)

	// Centered panel with rounded border
	panel := element.New(
		element.WithSize(30, 10),
		element.WithBorder(tui.BorderRounded),
		element.WithBorderStyle(tui.NewStyle().Foreground(tui.Cyan)),
		element.WithDirection(layout.Column),
		element.WithPadding(1),
		element.WithJustify(layout.JustifyCenter),
		element.WithAlign(layout.AlignCenter),
		element.WithGap(2),
	)

	// Title text - intrinsic width, centered by panel's AlignCenter
	title := element.NewText("Layout Engine Demo",
		element.WithTextStyle(tui.NewStyle().Foreground(tui.Green).Bold()),
	)

	// Hint text - intrinsic width, centered by panel's AlignCenter
	hint := element.NewText("Press any key to exit",
		element.WithTextStyle(tui.NewStyle().Foreground(tui.White)),
	)

	// Build the tree
	panel.AddChild(title.Element, hint.Element)
	root.AddChild(panel)

	// Create buffer once and reuse it
	buf := tui.NewBuffer(width, height)

	// Animation parameters
	minWidth := 30
	maxWidth := 60
	panelWidth := minWidth
	growing := true

	// Channel for keypress to exit
	done := make(chan struct{})
	go func() {
		b := make([]byte, 1)
		os.Stdin.Read(b)
		close(done)
	}()

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			// Update width - Yoga-style absolute float rounding prevents jitter
			if growing {
				panelWidth++
				if panelWidth >= maxWidth {
					growing = false
				}
			} else {
				panelWidth--
				if panelWidth <= minWidth {
					growing = true
				}
			}

			// Mutate the panel's style instead of rebuilding the tree
			style := panel.Style()
			style.Width = layout.Fixed(panelWidth)
			panel.SetStyle(style)

			// Clear buffer and re-render
			buf.Clear()
			root.Render(buf, width, height)

			// Render text elements
			element.RenderText(buf, title)
			element.RenderText(buf, hint)

			tui.Render(term, buf)
		}
	}
}
