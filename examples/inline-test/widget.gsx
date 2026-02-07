package main

import (
	"fmt"
	tui "github.com/grindlemire/go-tui"
)

type inlineApp struct {
	tuiApp  *tui.App
	counter int
	height  int
}

var _ tui.Component = (*inlineApp)(nil)
var _ tui.KeyListener = (*inlineApp)(nil)

func InlineApp(tuiApp *tui.App, height int) *inlineApp {
	return &inlineApp{
		tuiApp: tuiApp,
		height: height,
	}
}

func (a *inlineApp) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnRuneStop('p', func(ke tui.KeyEvent) { a.printLine() }),
		tui.OnRuneStop('q', func(ke tui.KeyEvent) { tui.Stop() }),
		tui.OnKeyStop(tui.KeyEscape, func(ke tui.KeyEvent) { tui.Stop() }),
	}
}

func (a *inlineApp) printLine() {
	a.counter++
	a.tuiApp.PrintAboveln("Line %d printed above the widget", a.counter)
}

templ (a *inlineApp) Render() {
	<div class="border-rounded" height={a.height}>
		<span>{fmt.Sprintf("Inline Widget | Counter: %d | Press 'p' to print, 'q' to quit", a.counter)}</span>
	</div>
}
