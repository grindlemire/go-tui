package main

import (
	"fmt"
	"os"

	tui "github.com/grindlemire/go-tui"
)

func main() {
	// Create a terminal
	term, err := tui.NewANSITerminal(os.Stdout, os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create terminal: %v\n", err)
		os.Exit(1)
	}

	// Enter raw mode and alt screen for clean display
	if err := term.EnterRawMode(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to enter raw mode: %v\n", err)
		os.Exit(1)
	}
	defer term.ExitRawMode()

	term.EnterAltScreen()
	defer term.ExitAltScreen()

	term.HideCursor()
	defer term.ShowCursor()

	// Get terminal size
	width, height := term.Size()

	// Create a buffer
	buf := tui.NewBuffer(width, height)

	// Define styles
	boxStyle := tui.NewStyle().Foreground(tui.Cyan)
	textStyle := tui.NewStyle().Foreground(tui.Green).Bold()
	hintStyle := tui.NewStyle().Foreground(tui.White)

	// Calculate box position (centered)
	boxWidth := 30
	boxHeight := 7
	boxX := (width - boxWidth) / 2
	boxY := (height - boxHeight) / 2

	// Draw a box with a title
	boxRect := tui.NewRect(boxX, boxY, boxWidth, boxHeight)
	tui.DrawBoxWithTitle(buf, boxRect, tui.BorderRounded, "Welcome", boxStyle)

	// Write text inside the box (centered)
	message := "Hello, World!"
	textX := boxX + (boxWidth-len(message))/2
	textY := boxY + boxHeight/2
	buf.SetString(textX, textY, message, textStyle)

	// Add hint below the box
	hint := "Press any key to exit"
	hintX := (width - len(hint)) / 2
	hintY := boxY + boxHeight + 1
	buf.SetString(hintX, hintY, hint, hintStyle)

	// Render to terminal
	tui.Render(term, buf)

	// Wait for any key press to exit
	b := make([]byte, 1)
	os.Stdin.Read(b)
}
