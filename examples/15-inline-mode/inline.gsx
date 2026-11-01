package main

import tui "github.com/grindlemire/go-tui"

type myApp struct {
	app          *tui.App
	showSettings *tui.State[bool]
	textarea     *tui.TextArea
}

func InlineApp() *myApp {
	a := &myApp{
		showSettings: tui.NewState(false),
	}
	a.textarea = tui.NewTextArea(
		tui.WithTextAreaWidth(60),
		tui.WithTextAreaBorder(tui.BorderRounded),
		tui.WithTextAreaPlaceholder("Type here..."),
		tui.WithTextAreaOnSubmit(a.send),
		tui.WithTextAreaAutoFocus(true),
	)
	return a
}

func (a *myApp) BindApp(app *tui.App) {
	a.app = app
	a.showSettings.BindApp(app)
	a.textarea.BindApp(app)
}

func (a *myApp) send(text string) {
	if text == "" {
		return
	}
	a.textarea.Clear()
	a.app.PrintAboveln("You: %s", text)
}

func (a *myApp) toggleSettings() {
	if a.showSettings.Get() {
		_ = a.app.ExitAlternateScreen()
		a.showSettings.Set(false)
		return
	}
	a.showSettings.Set(true)
	_ = a.app.EnterAlternateScreen()
}

func (a *myApp) KeyMap() tui.KeyMap {
	if a.showSettings.Get() {
		return tui.KeyMap{
			tui.On(tui.KeyEscape, func(ke tui.KeyEvent) { a.toggleSettings() }),
			tui.On(tui.Rune('c').Ctrl(), func(ke tui.KeyEvent) { ke.App().Stop() }),
		}
	}

	// The textarea is rendered with @a.textarea, which mounts it, so its own
	// KeyMap and Watchers are already active.
	return tui.KeyMap{
		tui.OnStop(tui.Rune('s').Ctrl(), func(ke tui.KeyEvent) { a.toggleSettings() }),
		tui.On(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.On(tui.Rune('c').Ctrl(), func(ke tui.KeyEvent) { ke.App().Stop() }),
	}
}

templ (a *myApp) Render() {
	if a.showSettings.Get() {
		<div class="flex-col h-full p-1 border-rounded border-cyan">
			<span class="font-bold text-cyan">Settings</span>
			<span class="font-dim">Press Escape to return</span>
		</div>
	} else {
		@a.textarea
	}
}
