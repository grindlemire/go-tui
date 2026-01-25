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
	app, err := tui.NewApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create app: %v\n", err)
		os.Exit(1)
	}
	defer app.Close()

	width, height := app.Size()

	// Root container
	root := element.New(
		element.WithSize(width, height),
		element.WithDirection(layout.Column),
	)

	// Header
	header := element.New(
		element.WithHeight(3),
		element.WithDirection(layout.Row),
		element.WithJustify(layout.JustifyCenter),
		element.WithAlign(layout.AlignCenter),
		element.WithBorder(tui.BorderSingle),
		element.WithBorderStyle(tui.NewStyle().Foreground(tui.Blue)),
	)
	headerTitle := element.New(
		element.WithText("Scrollable List Demo - Use Arrow Keys, j/k, PgUp/PgDn, Home/End"),
		element.WithTextStyle(tui.NewStyle().Foreground(tui.White).Bold()),
	)
	header.AddChild(headerTitle)

	// Main area with sidebar and content
	mainArea := element.New(
		element.WithFlexGrow(1),
		element.WithDirection(layout.Row),
	)

	// Scrollable list (sidebar) - now just an Element with WithScrollable!
	scrollableList := element.New(
		element.WithWidth(30),
		element.WithScrollable(element.ScrollVertical),
		element.WithDirection(layout.Column),
		element.WithBorder(tui.BorderSingle),
		element.WithBorderStyle(tui.NewStyle().Foreground(tui.Cyan)),
		element.WithPadding(1),
	)

	// Add many items to demonstrate scrolling
	for i := 0; i < 50; i++ {
		var style tui.Style
		if i%2 == 0 {
			style = tui.NewStyle().Foreground(tui.Green)
		} else {
			style = tui.NewStyle().Foreground(tui.Yellow)
		}

		item := element.New(
			element.WithText(fmt.Sprintf("Item %02d - Sample text", i+1)),
			element.WithTextStyle(style),
		)
		scrollableList.AddChild(item)
	}

	// Content area
	content := element.New(
		element.WithFlexGrow(1),
		element.WithDirection(layout.Column),
		element.WithJustify(layout.JustifyCenter),
		element.WithAlign(layout.AlignCenter),
		element.WithBorder(tui.BorderSingle),
		element.WithBorderStyle(tui.NewStyle().Foreground(tui.Magenta)),
	)

	instructions := element.New(
		element.WithText("Focus is on the scrollable list"),
		element.WithTextStyle(tui.NewStyle().Foreground(tui.White)),
	)
	content.AddChild(instructions)

	// No more .Element() unwrapping needed!
	mainArea.AddChild(scrollableList, content)

	// Footer with status
	footer := element.New(
		element.WithHeight(3),
		element.WithDirection(layout.Row),
		element.WithJustify(layout.JustifyCenter),
		element.WithAlign(layout.AlignCenter),
		element.WithBorder(tui.BorderSingle),
		element.WithBorderStyle(tui.NewStyle().Foreground(tui.Blue)),
	)
	footerText := element.New(
		element.WithText("Press ESC to exit"),
		element.WithTextStyle(tui.NewStyle().Foreground(tui.White)),
	)
	footer.AddChild(footerText)

	root.AddChild(header, mainArea, footer)
	app.SetRoot(root)

	// Register scrollable list for focus (auto-focuses first registered element)
	app.Focus().Register(scrollableList)

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
						scrollableList.ScrollBy(0, 1)
					case 'k':
						scrollableList.ScrollBy(0, -1)
					}
				}
				// Dispatch to focused element (the scrollable list)
				app.Dispatch(event)

			case tui.ResizeEvent:
				width, height = e.Width, e.Height
				style := root.Style()
				style.Width = layout.Fixed(width)
				style.Height = layout.Fixed(height)
				root.SetStyle(style)
				app.Dispatch(event)
			}
		}

		// Update footer with scroll position
		_, y := scrollableList.ScrollOffset()
		_, contentH := scrollableList.ContentSize()
		_, viewportH := scrollableList.ViewportSize()
		footerText.SetText(fmt.Sprintf("Scroll: %d/%d | Press ESC to exit", y, max(0, contentH-viewportH)))

		app.Render()
	}
}
