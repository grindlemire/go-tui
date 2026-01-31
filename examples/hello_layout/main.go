package main

import (
	"fmt"
	"os"
	"time"

	tui "github.com/grindlemire/go-tui"
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
	root := tui.New(
		tui.WithSize(width, height),
		tui.WithDirection(tui.Column),
		tui.WithJustify(tui.JustifyCenter),
		tui.WithAlign(tui.AlignCenter),
	)

	// Centered panel with rounded border
	panel := tui.New(
		tui.WithSize(30, 10),
		tui.WithBorder(tui.BorderRounded),
		tui.WithBorderStyle(tui.NewStyle().Foreground(tui.Cyan)),
		tui.WithDirection(tui.Column),
		tui.WithPadding(1),
		tui.WithJustify(tui.JustifyCenter),
		tui.WithAlign(tui.AlignCenter),
		tui.WithGap(2),
	)

	// Title text - intrinsic width, centered by panel's AlignCenter
	title := tui.New(
		tui.WithText("Layout Engine Demo"),
		tui.WithTextStyle(tui.NewStyle().Foreground(tui.Green).Bold()),
	)

	// Hint text - intrinsic width, centered by panel's AlignCenter
	hint := tui.New(
		tui.WithText("Press any key to exit"),
		tui.WithTextStyle(tui.NewStyle().Foreground(tui.White)),
	)

	// Build the tree
	panel.AddChild(title, hint)
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
			style.Width = tui.Fixed(panelWidth)
			panel.SetStyle(style)

			// Clear buffer and re-render
			buf.Clear()
			root.Render(buf, width, height)

			tui.Render(term, buf)
		}
	}
}
