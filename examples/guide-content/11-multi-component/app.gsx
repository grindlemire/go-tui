package main

import tui "github.com/grindlemire/go-tui"

type myApp struct {
	searchActive *tui.State[bool]
	query        *tui.State[string]
	category     *tui.State[string]
}

func MyApp() *myApp {
	return &myApp{
		searchActive: tui.NewState(false),
		query:        tui.NewState(""),
		category:     tui.NewState("Documents"),
	}
}

func (a *myApp) KeyMap() tui.KeyMap {
	km := tui.KeyMap{
		tui.OnKey(tui.KeyCtrlC, func(ke tui.KeyEvent) { ke.App().Stop() }),
	}
	if !a.searchActive.Get() {
		km = append(km, tui.OnRune('/', func(ke tui.KeyEvent) {
			a.searchActive.Set(true)
		}))
		km = append(km, tui.OnRune('q', func(ke tui.KeyEvent) {
			ke.App().Stop()
		}))
	}
	return km
}

templ (a *myApp) Render() {
	<div class="flex-col h-full border-rounded border-cyan">
		<div class="flex justify-center px-1 shrink-0">
			<span class="text-gradient-cyan-magenta font-bold">File Explorer</span>
		</div>
		<hr />
		<div class="flex grow min-h-0 overflow-hidden">
			@Sidebar(a.category)
			@Content(a.category, a.query)
		</div>
		@SearchBar(a.searchActive, a.query)
		<hr />
		<div class="flex justify-center px-1 shrink-0">
			<span class="font-dim">/search | Ctrl+B sidebar | j/k navigate | q quit</span>
		</div>
	</div>
}
