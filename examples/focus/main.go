// Package main demonstrates the focus management system.
// This example shows how to create focusable elements and navigate
// between them using Tab/Shift+Tab.
package main

import (
	"fmt"
	"os"
	"time"

	tui "github.com/grindlemire/go-tui"
)

func main() {
	// Create the application
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
		tui.WithPadding(2),
		tui.WithGap(2),
	)

	// Title
	title := tui.New(
		tui.WithText("Focus Navigation Demo"),
		tui.WithTextStyle(tui.NewStyle().Foreground(tui.White).Bold()),
	)

	// Instructions
	instructions := tui.New(
		tui.WithText("Press Tab/Shift+Tab to navigate, ESC to exit"),
		tui.WithTextStyle(tui.NewStyle().Foreground(tui.Cyan)),
	)

	// Container for focusable boxes
	boxContainer := tui.New(
		tui.WithFlexGrow(1),
		tui.WithDirection(tui.Row),
		tui.WithJustify(tui.JustifyCenter),
		tui.WithAlign(tui.AlignCenter),
		tui.WithGap(4),
	)

	// Create focusable boxes with different colors
	box1 := createBox("Red Box", tui.Red)
	box2 := createBox("Green Box", tui.Green)
	box3 := createBox("Blue Box", tui.Blue)
	box4 := createBox("Yellow Box", tui.Yellow)

	// Add boxes to container
	boxContainer.AddChild(box1, box2, box3, box4)

	// Status line at bottom
	statusLine := tui.New(
		tui.WithText(""),
		tui.WithTextStyle(tui.NewStyle().Foreground(tui.White)),
	)

	// Build the tree
	root.AddChild(title, instructions, boxContainer, statusLine)
	app.SetRoot(root)

	// Main event loop
	for {
		// Poll for events
		event, ok := app.PollEvent(50 * time.Millisecond)
		if ok {
			switch e := event.(type) {
			case tui.KeyEvent:
				switch e.Key {
				case tui.KeyEscape:
					return
				case tui.KeyTab:
					if e.Mod.Has(tui.ModShift) {
						app.FocusPrev()
					} else {
						app.FocusNext()
					}
				default:
					// Let focused element handle other keys
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

		// Update status line to show which box is focused
		focused := app.Focused()
		if focused != nil {
			if elem, ok := focused.(*tui.Element); ok {
				// Find the label child
				children := elem.Children()
				if len(children) > 0 && children[0].Text() != "" {
					statusLine.SetText(fmt.Sprintf("Currently focused: %s", children[0].Text()))
				}
			}
		}

		// Render
		app.Render()
	}
}

// createBox creates a focusable box with a label.
func createBox(label string, color tui.Color) *tui.Element {
	normalStyle := tui.NewStyle().Foreground(color)
	focusedStyle := tui.NewStyle().Foreground(tui.White).Background(color).Bold()

	// Create the label element
	labelElem := tui.New(
		tui.WithText(label),
		tui.WithTextStyle(tui.NewStyle().Foreground(tui.White)),
	)

	// Declare box first so closures can capture it
	var box *tui.Element
	box = tui.New(
		tui.WithSize(20, 5),
		tui.WithBorder(tui.BorderSingle),
		tui.WithBorderStyle(normalStyle),
		tui.WithDirection(tui.Column),
		tui.WithJustify(tui.JustifyCenter),
		tui.WithAlign(tui.AlignCenter),
		tui.WithOnFocus(func(el *tui.Element) {
			box.SetBorderStyle(focusedStyle)
			box.SetBorder(tui.BorderDouble)
			// Make label bold when focused
			labelElem.SetTextStyle(tui.NewStyle().Foreground(tui.White).Bold())
		}),
		tui.WithOnBlur(func(el *tui.Element) {
			box.SetBorderStyle(normalStyle)
			box.SetBorder(tui.BorderSingle)
			// Revert label style when blurred
			labelElem.SetTextStyle(tui.NewStyle().Foreground(tui.White))
		}),
	)

	box.AddChild(labelElem)
	return box
}
