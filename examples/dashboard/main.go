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
	// Create the application (handles terminal setup, raw mode, alternate screen)
	app, err := tui.NewApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create app: %v\n", err)
		os.Exit(1)
	}
	defer app.Close()

	width, height := app.Size()

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

	headerTitle := element.New(
		element.WithText("Dashboard"),
		element.WithTextStyle(tui.NewStyle().Foreground(tui.White).Bold()),
	)
	header.AddChild(headerTitle)

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

	sidebarTitle := element.New(
		element.WithText("Menu"),
		element.WithTextStyle(tui.NewStyle().Foreground(tui.Magenta).Bold()),
	)
	menuItem1 := element.New(
		element.WithText("> Overview"),
		element.WithTextStyle(tui.NewStyle().Foreground(tui.Green)),
	)
	menuItem2 := element.New(
		element.WithText("  Settings"),
		element.WithTextStyle(tui.NewStyle().Foreground(tui.White)),
	)
	menuItem3 := element.New(
		element.WithText("  Help"),
		element.WithTextStyle(tui.NewStyle().Foreground(tui.White)),
	)
	sidebar.AddChild(sidebarTitle, menuItem1, menuItem2, menuItem3)

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

	cardTitle := element.New(
		element.WithText("Status Card"),
		element.WithTextStyle(tui.NewStyle().Foreground(tui.Cyan).Bold()),
	)
	cardStatus := element.New(
		element.WithText("Systems Online"),
		element.WithTextStyle(tui.NewStyle().Foreground(tui.Green)),
	)
	cardHint := element.New(
		element.WithText("Press ESC to exit"),
		element.WithTextStyle(tui.NewStyle().Foreground(tui.White)),
	)

	card.AddChild(cardTitle, cardStatus, cardHint)
	content.AddChild(card)

	mainArea.AddChild(sidebar, content)

	// Build the tree
	root.AddChild(header, mainArea)
	app.SetRoot(root)

	// Animation parameters for the floating card
	minWidth := 25
	maxWidth := 50
	minHeight := 6
	maxHeight := 12
	cardWidth := minWidth
	cardHeight := minHeight
	growing := true

	// Main event loop using polling
	for {
		// Poll for events with a 50ms timeout (animation frame rate)
		event, ok := app.PollEvent(50 * time.Millisecond)
		if ok {
			switch e := event.(type) {
			case tui.KeyEvent:
				// Exit on Escape key
				if e.Key == tui.KeyEscape {
					return
				}
				// Also exit on any other key press for backward compatibility
				return
			case tui.ResizeEvent:
				// Handle resize: update root size and re-render
				width, height = e.Width, e.Height
				style := root.Style()
				style.Width = layout.Fixed(width)
				style.Height = layout.Fixed(height)
				root.SetStyle(style)
				app.Dispatch(event)
			}
		}

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
		cardStyle := card.Style()
		cardStyle.Width = layout.Fixed(cardWidth)
		cardStyle.Height = layout.Fixed(cardHeight)
		card.SetStyle(cardStyle)

		// Render using App
		app.Render()
	}
}
