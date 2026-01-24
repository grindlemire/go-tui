package main

import (
	"fmt"
	"os"

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

	// Simple layout: root container with a centered box
	root := element.New(
		element.WithSize(width, height),
		element.WithDirection(layout.Column),
		element.WithJustify(layout.JustifyCenter),
		element.WithAlign(layout.AlignCenter),
	)

	// Centered box - fixed size with border
	box := element.New(
		element.WithSize(40, 10),
		element.WithBorder(tui.BorderRounded),
		element.WithBorderStyle(tui.NewStyle().Foreground(tui.Cyan)),
	)

	root.AddChild(box)

	// Calculate layout
	root.Calculate(width, height)

	// Render
	buf := tui.NewBuffer(width, height)

	// Draw the box border
	rect := box.Rect()
	tui.DrawBox(buf, tui.NewRect(rect.X, rect.Y, rect.Width, rect.Height), box.Border(), box.BorderStyle())

	// Draw centered text
	msg := "Layout Engine Demo"
	msgX := rect.X + (rect.Width-len(msg))/2
	msgY := rect.Y + rect.Height/2
	textStyle := tui.NewStyle().Foreground(tui.Green).Bold()
	buf.SetString(msgX, msgY, msg, textStyle)

	hint := "Press any key to exit"
	hintX := rect.X + (rect.Width-len(hint))/2
	hintY := rect.Y + rect.Height/2 + 2
	buf.SetString(hintX, hintY, hint, tui.NewStyle().Foreground(tui.White))

	tui.Render(term, buf)

	b := make([]byte, 1)
	os.Stdin.Read(b)
}
