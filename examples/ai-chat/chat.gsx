package main

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	tui "github.com/grindlemire/go-tui"
	"github.com/grindlemire/go-tui/examples/ai-chat/settings"
)

type chat struct {
	app          *tui.App
	width        int
	textarea     *tui.TextArea
	showSettings *tui.State[bool]
	settingsView *settings.SettingsApp
	streaming    *tui.State[bool]
	eventCh      chan streamEvent
	lineBuf      strings.Builder
	cmd          *exec.Cmd
	firstMsg     bool
	// settings state (shared with settings view)
	model          *tui.State[string]
	maxTurns       *tui.State[int]
	permissionMode *tui.State[string]
	systemPrompt   *tui.State[string]
}

func Chat(width int) *chat {
	c := &chat{
		width:        width,
		showSettings: tui.NewState(false),
		streaming:    tui.NewState(false),
		eventCh:      make(chan streamEvent, 64),
		firstMsg:     true,
	}

	c.textarea = tui.NewTextArea(
		tui.WithTextAreaWidth(width-2), // -2 for border
		tui.WithTextAreaBorder(tui.BorderRounded),
		tui.WithTextAreaPlaceholder("Type a message..."),
		tui.WithTextAreaOnSubmit(c.submit),
	)

	c.model = tui.NewState("sonnet")
	c.maxTurns = tui.NewState(25)
	c.permissionMode = tui.NewState("default")
	c.systemPrompt = tui.NewState("You are a helpful assistant.")

	c.settingsView = settings.NewSettingsApp(
		c.model,
		c.maxTurns,
		c.permissionMode,
		c.systemPrompt,
		c.toggleSettings,
	)

	return c
}

func (c *chat) Init() func() {
	return func() {
		c.cancelStream()
	}
}

func (c *chat) submit(text string) {
	if text == "" || c.streaming.Get() {
		return
	}
	c.textarea.Clear()
	c.updateHeight()
	c.app.PrintAboveln("You: %s", text)
	c.streaming.Set(true)
	c.startClaude(text)
}

func (c *chat) startClaude(message string) {
	args := []string{"-p", message, "--output-format", "stream-json", "--verbose", "--include-partial-messages"}

	if !c.firstMsg {
		args = append(args, "--continue")
	}
	c.firstMsg = false

	args = append(args, "--model", c.model.Get())

	if mt := c.maxTurns.Get(); mt > 0 {
		args = append(args, "--max-turns", fmt.Sprintf("%d", mt))
	}

	if pm := c.permissionMode.Get(); pm != "default" {
		args = append(args, "--permission-mode", pm)
	}

	if sp := c.systemPrompt.Get(); sp != "" {
		args = append(args, "--append-system-prompt", sp)
	}

	cmd := exec.Command("claude", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		c.eventCh <- streamEvent{Type: eventError, Text: fmt.Sprintf("pipe error: %v", err)}
		return
	}

	if err := cmd.Start(); err != nil {
		c.eventCh <- streamEvent{Type: eventError, Text: fmt.Sprintf("start error: %v", err)}
		return
	}
	c.cmd = cmd

	go func() {
		parser := newStreamParser()
		scanner := bufio.NewScanner(stdout)
		scanner.Buffer(make([]byte, 0, 256*1024), 256*1024)
		for scanner.Scan() {
			for _, ev := range parser.parseLine(scanner.Bytes()) {
				c.eventCh <- ev
			}
		}
		cmd.Wait()
		c.eventCh <- streamEvent{Type: eventDone}
	}()
}

var agentGradient = tui.NewGradient(tui.BrightCyan, tui.BrightMagenta)

func gradientLine(text string) string {
	runes := []rune(text)
	if len(runes) == 0 {
		return ""
	}

	var b strings.Builder
	b.Grow(len(text) * 20) // ANSI overhead per char
	for i, r := range runes {
		t := float64(i) / float64(max(len(runes)-1, 1))
		cr, cg, cb := agentGradient.At(t).ToRGBValues()
		fmt.Fprintf(&b, "\033[38;2;%d;%d;%dm%c", cr, cg, cb, r)
	}
	b.WriteString("\033[0m")
	return b.String()
}

func (c *chat) onStreamEvent(ev streamEvent) {
	switch ev.Type {
	case eventText:
		// Accumulate text, flush complete lines
		c.lineBuf.WriteString(ev.Text)
		for {
			s := c.lineBuf.String()
			idx := strings.Index(s, "\n")
			if idx < 0 {
				break
			}
			c.app.PrintAboveStyledln("%s", gradientLine(s[:idx]))
			c.lineBuf.Reset()
			c.lineBuf.WriteString(s[idx+1:])
		}

	case eventToolUse:
		// Flush any pending text first
		if c.lineBuf.Len() > 0 {
			c.app.PrintAboveStyledln("%s", gradientLine(c.lineBuf.String()))
			c.lineBuf.Reset()
		}
		c.app.PrintAboveStyledln("%s", gradientLine("  > "+ev.Text))

	case eventError:
		if c.lineBuf.Len() > 0 {
			c.app.PrintAboveStyledln("%s", gradientLine(c.lineBuf.String()))
			c.lineBuf.Reset()
		}
		c.app.PrintAboveln("Error: %s", ev.Text)
		c.streaming.Set(false)
		c.cmd = nil

	case eventDone:
		if c.lineBuf.Len() > 0 {
			c.app.PrintAboveStyledln("%s", gradientLine(c.lineBuf.String()))
			c.lineBuf.Reset()
		}
		c.app.PrintAboveln("")
		c.streaming.Set(false)
		c.cmd = nil
	}
}

func (c *chat) cancelStream() {
	if c.cmd != nil && c.cmd.Process != nil {
		c.cmd.Process.Signal(syscall.SIGTERM)
	}
}

func (c *chat) toggleSettings() {
	if c.streaming.Get() {
		return
	}
	if c.showSettings.Get() {
		_ = c.app.ExitAlternateScreen()
		c.showSettings.Set(false)
		c.updateHeight()
		return
	}
	c.showSettings.Set(true)
	_ = c.app.EnterAlternateScreen()
}

func (c *chat) KeyMap() tui.KeyMap {
	if c.showSettings.Get() {
		km := c.settingsView.KeyMap()
		km = append(km,
			tui.OnKey(tui.KeyCtrlC, func(ke tui.KeyEvent) { ke.App().Stop() }),
		)
		return km
	}

	// While streaming, only allow Ctrl+C (kills subprocess) and Escape
	if c.streaming.Get() {
		return tui.KeyMap{
			tui.OnKeyStop(tui.KeyCtrlC, func(ke tui.KeyEvent) { c.cancelStream(); ke.App().Stop() }),
			tui.OnKeyStop(tui.KeyEscape, func(ke tui.KeyEvent) { c.cancelStream() }),
		}
	}

	km := c.textarea.KeyMap()
	km = append(km,
		tui.OnKeyStop(tui.KeyCtrlS, func(ke tui.KeyEvent) { c.toggleSettings() }),
		tui.OnKeyStop(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnKey(tui.KeyCtrlC, func(ke tui.KeyEvent) { ke.App().Stop() }),
	)
	return km
}

func (c *chat) Watchers() []tui.Watcher {
	w := c.textarea.Watchers()
	w = append(w, tui.NewChannelWatcher(c.eventCh, c.onStreamEvent))
	return w
}

func (c *chat) updateHeight() {
	h := c.textarea.Height()
	if h < 3 {
		h = 3
	}
	c.app.SetInlineHeight(h)
}

templ (c *chat) Render() {
	@if c.showSettings.Get() {
		@c.settingsView
	} @else {
		c.updateHeight()
		@c.textarea
	}
}
