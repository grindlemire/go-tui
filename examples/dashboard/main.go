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

	// Root container - full screen
	root := element.New(
		element.WithSize(width, height),
		element.WithDirection(layout.Column),
	)

	// Top container (header) - fixed height
	header := element.New(
		element.WithHeight(3),
		element.WithFlexGrow(0),
		element.WithDirection(layout.Row),
		element.WithJustify(layout.JustifyCenter),
		element.WithAlign(layout.AlignCenter),
		element.WithBorder(tui.BorderSingle),
		element.WithBorderStyle(tui.NewStyle().Foreground(tui.Blue)),
	)

	headerTitle := element.NewText("Dashboard",
		element.WithTextStyle(tui.NewStyle().Foreground(tui.White).Bold()),
	)
	header.AddChild(headerTitle.Element)

	// Main content area - row with sidebar and main
	mainArea := element.New(
		element.WithFlexGrow(1),
		element.WithDirection(layout.Row),
	)

	// Side container (sidebar) - fixed width
	sidebar := element.New(
		element.WithWidth(20),
		element.WithFlexGrow(0),
		element.WithDirection(layout.Column),
		element.WithPadding(1),
		element.WithGap(1),
		element.WithBorder(tui.BorderSingle),
		element.WithBorderStyle(tui.NewStyle().Foreground(tui.Magenta)),
	)

	sidebarTitle := element.NewText("Menu",
		element.WithTextStyle(tui.NewStyle().Foreground(tui.Magenta).Bold()),
	)
	menuItem1 := element.NewText("> Overview",
		element.WithTextStyle(tui.NewStyle().Foreground(tui.Green)),
	)
	menuItem2 := element.NewText("  Settings",
		element.WithTextStyle(tui.NewStyle().Foreground(tui.White)),
	)
	menuItem3 := element.NewText("  Help",
		element.WithTextStyle(tui.NewStyle().Foreground(tui.White)),
	)
	sidebar.AddChild(sidebarTitle.Element, menuItem1.Element, menuItem2.Element, menuItem3.Element)

	// Main content - fills remaining space with centered floating card
	content := element.New(
		element.WithFlexGrow(1),
		element.WithDirection(layout.Column),
		element.WithJustify(layout.JustifyCenter),
		element.WithAlign(layout.AlignCenter),
	)

	// Floating card that will animate
	card := element.New(
		element.WithSize(30, 8),
		element.WithBorder(tui.BorderRounded),
		element.WithBorderStyle(tui.NewStyle().Foreground(tui.Cyan)),
		element.WithDirection(layout.Column),
		element.WithPadding(1),
		element.WithJustify(layout.JustifyCenter),
		element.WithAlign(layout.AlignCenter),
		element.WithGap(1),
	)

	cardTitle := element.NewText("Status Card",
		element.WithTextStyle(tui.NewStyle().Foreground(tui.Cyan).Bold()),
	)
	cardStatus := element.NewText("Systems Online",
		element.WithTextStyle(tui.NewStyle().Foreground(tui.Green)),
	)
	cardHint := element.NewText("Press any key to exit",
		element.WithTextStyle(tui.NewStyle().Foreground(tui.White)),
	)

	card.AddChild(cardTitle.Element, cardStatus.Element, cardHint.Element)
	content.AddChild(card)

	mainArea.AddChild(sidebar, content)

	// Build the tree
	root.AddChild(header, mainArea)

	// Create buffer
	buf := tui.NewBuffer(width, height)

	// Animation parameters for the floating card
	minWidth := 25
	maxWidth := 50
	minHeight := 6
	maxHeight := 12
	cardWidth := minWidth
	cardHeight := minHeight
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
			// Update card dimensions - animate expand/contract
			if growing {
				cardWidth++
				cardHeight = minHeight + (cardWidth-minWidth)*(maxHeight-minHeight)/(maxWidth-minWidth)
				if cardWidth >= maxWidth {
					growing = false
				}
			} else {
				cardWidth--
				cardHeight = minHeight + (cardWidth-minWidth)*(maxHeight-minHeight)/(maxWidth-minWidth)
				if cardWidth <= minWidth {
					growing = true
				}
			}

			// Update card style
			style := card.Style()
			style.Width = layout.Fixed(cardWidth)
			style.Height = layout.Fixed(cardHeight)
			card.SetStyle(style)

			// Clear buffer and re-render
			buf.Clear()
			root.Render(buf, width, height)

			// Render text elements
			element.RenderText(buf, headerTitle)
			element.RenderText(buf, sidebarTitle)
			element.RenderText(buf, menuItem1)
			element.RenderText(buf, menuItem2)
			element.RenderText(buf, menuItem3)
			element.RenderText(buf, cardTitle)
			element.RenderText(buf, cardStatus)
			element.RenderText(buf, cardHint)

			tui.Render(term, buf)
		}
	}
}
