package main

import (
	"os"
	tui "github.com/grindlemire/go-tui"
)

type interactiveApp struct {
	events *tui.Events[string]
}

func Interactive() *interactiveApp {
	return &interactiveApp{
		events: tui.NewEvents[string](),
	}
}

func (a *interactiveApp) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnRune('q', func(ke tui.KeyEvent) { tui.Stop() }),
		tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { tui.Stop() }),
		tui.OnRune('p', func(ke tui.KeyEvent) {
			frame := tui.SnapshotFrame()
			os.WriteFile("debug-frame.txt", []byte(frame), 0644)
		}),
	}
}

templ (a *interactiveApp) Render() {
	<div class="flex-col p-1 border-rounded gap-1">
		<div class="flex justify-between shrink-0">
			<span class="text-gradient-cyan-magenta font-bold shrink-0">{"Interactive Elements"}</span>
			<span class="font-dim" minWidth={0}>{"[+/-/0/d] counter  [space/r] timer  [1/2/3] toggles  [p] print  [q] quit"}</span>
		</div>
		<div class="flex gap-1 items-stretch flex-1">
			@Counter(a.events)
			@Timer(a.events)
		</div>
		<div class="flex gap-1 items-stretch flex-1">
			@Toggles(a.events)
			@EventInspector(a.events)
		</div>
	</div>
}
