package main

import (
	"fmt"
	"os"
	"time"

	tui "github.com/grindlemire/go-tui"
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
		tui.WithText("Scrollable List Demo - Use Arrow Keys, j/k, PgUp/PgDn, Home/End"),
		tui.WithTextStyle(tui.NewStyle().Foreground(tui.White).Bold()),
	)
	header.AddChild(headerTitle)

	// Main area with sidebar and content
	mainArea := tui.New(
		tui.WithFlexGrow(1),
		tui.WithDirection(tui.Row),
	)

	// Scrollable list (sidebar) - now just an Element with WithScrollable!
	scrollableList := tui.New(
		tui.WithWidth(30),
		tui.WithScrollable(tui.ScrollVertical),
		tui.WithDirection(tui.Column),
		tui.WithBorder(tui.BorderSingle),
		tui.WithBorderStyle(tui.NewStyle().Foreground(tui.Cyan)),
		tui.WithPadding(1),
	)

	// Add many items to demonstrate scrolling
	for i := 0; i < 50; i++ {
		var style tui.Style
		if i%2 == 0 {
			style = tui.NewStyle().Foreground(tui.Green)
		} else {
			style = tui.NewStyle().Foreground(tui.Yellow)
		}

		item := tui.New(
			tui.WithText(fmt.Sprintf("Item %02d - Sample text", i+1)),
			tui.WithTextStyle(style),
		)
		scrollableList.AddChild(item)
	}

	// Content area
	content := tui.New(
		tui.WithFlexGrow(1),
		tui.WithDirection(tui.Column),
		tui.WithJustify(tui.JustifyCenter),
		tui.WithAlign(tui.AlignCenter),
		tui.WithBorder(tui.BorderSingle),
		tui.WithBorderStyle(tui.NewStyle().Foreground(tui.Magenta)),
	)

	instructions := tui.New(
		tui.WithText("Focus is on the scrollable list"),
		tui.WithTextStyle(tui.NewStyle().Foreground(tui.White)),
	)
	content.AddChild(instructions)

	// No more .Element() unwrapping needed!
	mainArea.AddChild(scrollableList, content)

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
				style.Width = tui.Fixed(width)
				style.Height = tui.Fixed(height)
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
