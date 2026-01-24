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

	// Build layout tree with Element API - much cleaner than before!
	root := element.New(
		element.WithSize(width, height),
		element.WithDirection(layout.Column),
		element.WithJustify(layout.JustifyCenter),
		element.WithAlign(layout.AlignCenter),
	)

	// Centered panel with rounded border
	panel := element.New(
		element.WithSize(40, 10),
		element.WithBorder(tui.BorderRounded),
		element.WithBorderStyle(tui.NewStyle().Foreground(tui.Cyan)),
		element.WithDirection(layout.Column),
		element.WithPadding(1),
		element.WithJustify(layout.JustifyCenter),
		element.WithAlign(layout.AlignCenter),
		element.WithGap(2),
	)

	// Title text
	title := element.NewText("Layout Engine Demo",
		element.WithTextStyle(tui.NewStyle().Foreground(tui.Green).Bold()),
		element.WithTextAlign(element.TextAlignCenter),
		element.WithElementOption(element.WithHeight(1)),
	)

	// Hint text
	hint := element.NewText("Press any key to exit",
		element.WithTextStyle(tui.NewStyle().Foreground(tui.White)),
		element.WithTextAlign(element.TextAlignCenter),
		element.WithElementOption(element.WithHeight(1)),
	)

	// Build the tree
	panel.AddChild(title.Element, hint.Element)
	root.AddChild(panel)

	// Calculate layout and render
	buf := tui.NewBuffer(width, height)
	root.Render(buf, width, height)

	// Render text elements (Text content requires explicit rendering)
	element.RenderText(buf, title)
	element.RenderText(buf, hint)

	tui.Render(term, buf)

	// Wait for keypress
	b := make([]byte, 1)
	os.Stdin.Read(b)
}
