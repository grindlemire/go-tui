package main

import tui "github.com/grindlemire/go-tui"

type interactiveApp struct{}

func Interactive() *interactiveApp {
	return &interactiveApp{}
}

func (a *interactiveApp) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnRune('q', func(ke tui.KeyEvent) { tui.Stop() }),
		tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { tui.Stop() }),
	}
}

templ (a *interactiveApp) Render() {
	<div class="flex-col p-1 border-rounded gap-1">
		<div class="flex justify-between">
			<span class="text-gradient-cyan-magenta font-bold">{"Interactive Elements"}</span>
			<span class="font-dim">{"[q] quit"}</span>
		</div>
		<div class="flex gap-1">
			@Counter()
			@Timer()
		</div>
		<div class="flex gap-1">
			@Toggles()
		</div>
	</div>
}
