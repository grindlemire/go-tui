package main

import (
	tui "github.com/grindlemire/go-tui"
)

type chat struct {
	width    int
	textarea *tui.TextArea
}

func Chat(width int) *chat {
	c := &chat{
		width: width,
	}
	c.textarea = tui.NewTextArea(
		tui.WithTextAreaWidth(width-2), // -2 for border
		tui.WithTextAreaBorder(tui.BorderRounded),
		tui.WithTextAreaPlaceholder("Type a message..."),
		tui.WithTextAreaOnSubmit(c.submit),
	)
	return c
}

func (c *chat) submit(text string) {
	if text == "" {
		return
	}
	tui.PrintAboveln("You: %s", text)
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

func (c *chat) updateHeight() {
	h := c.textarea.Height()
	if h < 3 {
		h = 3
	}
	tui.SetInlineHeight(h)
}

templ (c *chat) Render() {
	c.updateHeight()
	@c.textarea
}
