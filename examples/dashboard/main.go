package main

import (
	"fmt"
	"os"
	"time"

	tui "github.com/grindlemire/go-tui"
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
	root := tui.New(
		tui.WithSize(width, height),
		tui.WithDirection(tui.Column),
	)

	// Top container (header) - fixed height
	header := tui.New(
		tui.WithHeight(3),
		tui.WithFlexGrow(0),
		tui.WithDirection(tui.Row),
		tui.WithJustify(tui.JustifyCenter),
		tui.WithAlign(tui.AlignCenter),
		tui.WithBorder(tui.BorderSingle),
		tui.WithBorderStyle(tui.NewStyle().Foreground(tui.Blue)),
	)

	headerTitle := tui.New(
		tui.WithText("Dashboard"),
		tui.WithTextStyle(tui.NewStyle().Foreground(tui.White).Bold()),
	)
	header.AddChild(headerTitle)

	// Main content area - row with sidebar and main
	mainArea := tui.New(
		tui.WithFlexGrow(1),
		tui.WithDirection(tui.Row),
	)

	// Side container (sidebar) - fixed width
	sidebar := tui.New(
		tui.WithWidth(20),
		tui.WithFlexGrow(0),
		tui.WithDirection(tui.Column),
		tui.WithPadding(1),
		tui.WithGap(1),
		tui.WithBorder(tui.BorderSingle),
		tui.WithBorderStyle(tui.NewStyle().Foreground(tui.Magenta)),
	)

	sidebarTitle := tui.New(
		tui.WithText("Menu"),
		tui.WithTextStyle(tui.NewStyle().Foreground(tui.Magenta).Bold()),
	)
	menuItem1 := tui.New(
		tui.WithText("> Overview"),
		tui.WithTextStyle(tui.NewStyle().Foreground(tui.Green)),
	)
	menuItem2 := tui.New(
		tui.WithText("  Settings"),
		tui.WithTextStyle(tui.NewStyle().Foreground(tui.White)),
	)
	menuItem3 := tui.New(
		tui.WithText("  Help"),
		tui.WithTextStyle(tui.NewStyle().Foreground(tui.White)),
	)
	sidebar.AddChild(sidebarTitle, menuItem1, menuItem2, menuItem3)

	// Main content - fills remaining space with centered floating card
	content := tui.New(
		tui.WithFlexGrow(1),
		tui.WithDirection(tui.Column),
		tui.WithJustify(tui.JustifyCenter),
		tui.WithAlign(tui.AlignCenter),
	)

	// Floating card that will animate
	card := tui.New(
		tui.WithSize(30, 8),
		tui.WithBorder(tui.BorderRounded),
		tui.WithBorderStyle(tui.NewStyle().Foreground(tui.Cyan)),
		tui.WithDirection(tui.Column),
		tui.WithPadding(1),
		tui.WithJustify(tui.JustifyCenter),
		tui.WithAlign(tui.AlignCenter),
		tui.WithGap(1),
	)

	cardTitle := tui.New(
		tui.WithText("Status Card"),
		tui.WithTextStyle(tui.NewStyle().Foreground(tui.Cyan).Bold()),
	)
	cardStatus := tui.New(
		tui.WithText("Systems Online"),
		tui.WithTextStyle(tui.NewStyle().Foreground(tui.Green)),
	)
	cardHint := tui.New(
		tui.WithText("Press ESC to exit"),
		tui.WithTextStyle(tui.NewStyle().Foreground(tui.White)),
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
				style.Width = tui.Fixed(width)
				style.Height = tui.Fixed(height)
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
		cardStyle.Width = tui.Fixed(cardWidth)
		cardStyle.Height = tui.Fixed(cardHeight)
		card.SetStyle(cardStyle)

		// Render using App
		app.Render()
	}
}
