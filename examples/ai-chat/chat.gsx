package main

import (
	tui "github.com/grindlemire/go-tui"
	"github.com/grindlemire/go-tui/internal/debug"
)

type chat struct {
	app      *tui.App
	width    int
	textarea *TextArea
}

func Chat(app *tui.App, width int) *chat {
	c := &chat{
		app:   app,
		width: width,
	}
	c.textarea = NewTextArea(width-2, c.submit) // -2 for border
	return c
}

func (c *chat) submit(text string) {
	if text == "" {
		return
	}
	c.app.PrintAboveln("You: %s", text)
	c.textarea.Clear()
}

func (c *chat) KeyMap() tui.KeyMap {
	km := c.textarea.KeyMap()
	// Add quit keys
	km = append(km,
		tui.OnKeyStop(tui.KeyEscape, func(ke tui.KeyEvent) { tui.Stop() }),
		tui.OnKey(tui.KeyCtrlC, func(ke tui.KeyEvent) { tui.Stop() }),
	)
	return km
}

func (c *chat) Watchers() []tui.Watcher {
	return c.textarea.Watchers()
}

templ (c *chat) Render() {
	<div class="border-rounded flex-col" height={c.totalHeight()}>
		@for i, _ := range c.textarea.Lines() {
			<span>{c.textarea.LineWithCursor(i)}</span>
		}
	</div>
}

func (c *chat) totalHeight() int {
	h := c.textarea.Height()
	debug.Log("chat.totalHeight: textarea.Height()=%d", h)
	if h < 1 {
		h = 1
	}
	// Total height including border
	total := h + 2
	if total < 3 {
		total = 3
	}
	if total > 10 {
		total = 10
	}
	debug.Log("chat.totalHeight: calling SetInlineHeight(%d)", total)
	c.app.SetInlineHeight(total)
	return total
}
