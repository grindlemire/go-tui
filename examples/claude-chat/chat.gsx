package main

import (
	"bufio"
	"os/exec"
	"strings"

	"github.com/grindlemire/go-tui/pkg/tui"
)

// ============================================================================
// Layout Helpers
// ============================================================================

// WrapText splits text into lines that fit within width
func WrapText(text string, width int) []string {
	if width <= 0 || len(text) == 0 {
		return []string{text}
	}

	var lines []string
	var current strings.Builder
	currentWidth := 0

	for _, r := range text {
		runeWidth := tui.RuneWidth(r)
		if currentWidth+runeWidth > width && currentWidth > 0 {
			lines = append(lines, current.String())
			current.Reset()
			currentWidth = 0
		}
		current.WriteRune(r)
		currentWidth += runeWidth
	}
	if current.Len() > 0 {
		lines = append(lines, current.String())
	}
	return lines
}

// ============================================================================
// Claude CLI Integration
// ============================================================================

// SendToClaudeAndStream calls claude CLI and streams response via PrintAboveln
func SendToClaudeAndStream(app *tui.App, query string) {
	// Print user query
	app.PrintAboveln("You: %s", query)
	app.PrintAboveln("")

	// Call claude CLI with print mode
	cmd := exec.Command("claude", "-p", query)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		app.PrintAboveln("Error: %v", err)
		return
	}
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		app.PrintAboveln("Error starting claude: %v", err)
		return
	}

	// Stream response line by line
	scanner := bufio.NewScanner(stdout)
	firstLine := true
	for scanner.Scan() {
		line := scanner.Text()
		if firstLine {
			app.PrintAboveln("Claude: %s", line)
			firstLine = false
		} else {
			app.PrintAboveln("        %s", line)
		}
	}

	// Check for errors
	errScanner := bufio.NewScanner(stderr)
	for errScanner.Scan() {
		app.PrintAboveln("Error: %s", errScanner.Text())
	}

	cmd.Wait()
	app.PrintAboveln("")
}

// ============================================================================
// Event Handlers
// ============================================================================

// CreateKeyHandler returns a key handler for the chat input
func CreateKeyHandler(buf *TextBuffer, app *tui.App, updateView func()) func(tui.KeyEvent) bool {
	return func(e tui.KeyEvent) bool {
		switch e.Key {
		case tui.KeyRune:
			buf.Insert(e.Rune)
		case tui.KeyBackspace:
			buf.Backspace()
		case tui.KeyDelete:
			buf.Delete()
		case tui.KeyLeft:
			buf.Left()
		case tui.KeyRight:
			buf.Right()
		case tui.KeyHome:
			buf.Home()
		case tui.KeyEnd:
			buf.End()
		case tui.KeyEnter:
			query := strings.TrimSpace(buf.String())
			if query != "" {
				buf.Clear()
				go SendToClaudeAndStream(app, query)
			}
		case tui.KeyEscape:
			app.Stop()
			return true
		default:
			return false
		}
		updateView()
		tui.MarkDirty()
		return true
	}
}

// ============================================================================
// Template Definition
// ============================================================================

// ChatInput renders the input box with wrapped text lines
templ ChatInput(lines []string) {
	<div class="border-rounded p-1 flex-col">
		@for _, line := range lines {
			<span>{line}</span>
		}
		@if len(lines) == 0 {
			<span>{"> \u2588"}</span>
		}
	</div>
}
