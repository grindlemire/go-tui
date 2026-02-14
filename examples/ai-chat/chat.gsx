package main

import (
	tui "github.com/grindlemire/go-tui"
	"github.com/grindlemire/go-tui/examples/ai-chat/settings"
)

type chat struct {
	app          *tui.App
	width        int
	textarea     *tui.TextArea
	showSettings *tui.State[bool]
	settingsView *settings.SettingsApp
}

var _ tui.AppBinder = (*chat)(nil)

func (c *chat) BindApp(app *tui.App) {
	c.app = app
	c.showSettings.BindApp(app)
	c.textarea.BindApp(app)
}

func Chat(width int) *chat {
	c := &chat{
		width:        width,
		showSettings: tui.NewState(false),
	}

	c.textarea = tui.NewTextArea(
		tui.WithTextAreaWidth(width-2), // -2 for border
		tui.WithTextAreaBorder(tui.BorderRounded),
		tui.WithTextAreaPlaceholder("Type a message..."),
		tui.WithTextAreaOnSubmit(c.submit),
	)

	provider := tui.NewState("openai")
	model := tui.NewState("gpt-4.1-mini")
	temperature := tui.NewState(0.7)
	systemPrompt := tui.NewState("You are a helpful assistant.")
	availableProviders := []string{"openai", "anthropic", "ollama"}
	providerModels := map[string][]string{
		"openai":    {"gpt-4.1-mini", "gpt-4.1", "gpt-4o-mini"},
		"anthropic": {"claude-3-5-haiku-latest", "claude-3-7-sonnet-latest"},
		"ollama":    {"llama3.2", "mistral", "codellama"},
	}
	c.settingsView = settings.NewSettingsApp(
		provider,
		model,
		temperature,
		systemPrompt,
		availableProviders,
		providerModels,
		c.toggleSettings,
	)

	return c
}

func (c *chat) submit(text string) {
	if text == "" {
		return
	}
	c.textarea.Clear()
	c.updateHeight()
	c.app.PrintAboveln("You: %s", text)
}

func (c *chat) toggleSettings() {
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

	km := c.textarea.KeyMap()
	km = append(km,
		tui.OnKeyStop(tui.KeyCtrlS, func(ke tui.KeyEvent) { c.toggleSettings() }),
		tui.OnKeyStop(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnKey(tui.KeyCtrlC, func(ke tui.KeyEvent) { ke.App().Stop() }),
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
