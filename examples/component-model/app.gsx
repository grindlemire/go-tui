package main

import tui "github.com/grindlemire/go-tui"

type myApp struct {
	searchActive *tui.State[bool]
	query        *tui.State[string]
}

func MyApp() *myApp {
	return &myApp{
		searchActive: tui.NewState(false),
		query:        tui.NewState(""),
	}
}

func (a *myApp) KeyMap() tui.KeyMap {
	km := tui.KeyMap{
		tui.OnKey(tui.KeyCtrlC, func(ke tui.KeyEvent) { tui.Stop() }),
	}
	if !a.searchActive.Get() {
		km = append(km, tui.OnRune('/', func(ke tui.KeyEvent) {
			a.searchActive.Set(true)
		}))
	}
	return km
}

templ (a *myApp) Render() {
	<div class="flex">
		@Sidebar(a.query)
		<div class="flex-col flex-grow-1">
			@SearchInput(a.searchActive, a.query)
			<div class="flex-grow-1 p-1">
				<span>Press / to search. Ctrl+B toggles sidebar. Ctrl+C quits.</span>
			</div>
		</div>
	</div>
}
