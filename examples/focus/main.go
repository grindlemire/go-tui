// Package main demonstrates the focus management system.
// This example shows how to create focusable elements and navigate
// between them using Tab/Shift+Tab.
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
	// Create the application
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
		element.WithPadding(2),
		element.WithGap(2),
	)

	// Title
	title := element.New(
		element.WithText("Focus Navigation Demo"),
		element.WithTextStyle(tui.NewStyle().Foreground(tui.White).Bold()),
	)

	// Instructions
	instructions := element.New(
		element.WithText("Press Tab/Shift+Tab to navigate, ESC to exit"),
		element.WithTextStyle(tui.NewStyle().Foreground(tui.Cyan)),
	)

	// Container for focusable boxes
	boxContainer := element.New(
		element.WithFlexGrow(1),
		element.WithDirection(layout.Row),
		element.WithJustify(layout.JustifyCenter),
		element.WithAlign(layout.AlignCenter),
		element.WithGap(4),
	)

	// Create focusable boxes with different colors
	box1 := createBox("Red Box", tui.Red)
	box2 := createBox("Green Box", tui.Green)
	box3 := createBox("Blue Box", tui.Blue)
	box4 := createBox("Yellow Box", tui.Yellow)

	// Add boxes to container
	boxContainer.AddChild(box1, box2, box3, box4)

	// Status line at bottom
	statusLine := element.New(
		element.WithText(""),
		element.WithTextStyle(tui.NewStyle().Foreground(tui.White)),
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
				style.Width = layout.Fixed(width)
				style.Height = layout.Fixed(height)
				root.SetStyle(style)
				app.Dispatch(event)
			}
		}

		// Update status line to show which box is focused
		focused := app.Focused()
		if focused != nil {
			if elem, ok := focused.(*element.Element); ok {
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
func createBox(label string, color tui.Color) *element.Element {
	normalStyle := tui.NewStyle().Foreground(color)
	focusedStyle := tui.NewStyle().Foreground(tui.White).Background(color).Bold()

	// Create the label element
	labelElem := element.New(
		element.WithText(label),
		element.WithTextStyle(tui.NewStyle().Foreground(tui.White)),
	)

	// Declare box first so closures can capture it
	var box *element.Element
	box = element.New(
		element.WithSize(20, 5),
		element.WithBorder(tui.BorderSingle),
		element.WithBorderStyle(normalStyle),
		element.WithDirection(layout.Column),
		element.WithJustify(layout.JustifyCenter),
		element.WithAlign(layout.AlignCenter),
		element.WithOnFocus(func() {
			box.SetBorderStyle(focusedStyle)
			box.SetBorder(tui.BorderDouble)
			// Make label bold when focused
			labelElem.SetTextStyle(tui.NewStyle().Foreground(tui.White).Bold())
		}),
		element.WithOnBlur(func() {
			box.SetBorderStyle(normalStyle)
			box.SetBorder(tui.BorderSingle)
			// Revert label style when blurred
			labelElem.SetTextStyle(tui.NewStyle().Foreground(tui.White))
		}),
	)

	box.AddChild(labelElem)
	return box
}
