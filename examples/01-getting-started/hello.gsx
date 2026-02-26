package main

import tui "github.com/grindlemire/go-tui"

type helloApp struct{}

func Hello() *helloApp {
	return &helloApp{}
}

func (h *helloApp) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnRune('q', func(ke tui.KeyEvent) { ke.App().Stop() }),
	}
}

templ (h *helloApp) Render() {
	<div class="flex-col items-center justify-center h-full">
		<div class="border-rounded border-cyan p-2 gap-1 flex-col items-center">
			<span class="text-cyan font-bold">Hello, Terminal!</span>
			<br />
			<span class="font-dim">Press q to quit</span>
		</div>
	</div>
}
