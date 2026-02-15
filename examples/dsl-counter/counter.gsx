package main

import (
	"fmt"
	tui "github.com/grindlemire/go-tui"
)

type counterApp struct {
	count *tui.State[int]
}

func Counter() *counterApp {
	return &counterApp{
		count: tui.NewState(0),
	}
}

func (c *counterApp) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnRune('q', func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnRune('+', func(ke tui.KeyEvent) { c.count.Set(c.count.Get() + 1) }),
		tui.OnRune('=', func(ke tui.KeyEvent) { c.count.Set(c.count.Get() + 1) }),
		tui.OnRune('-', func(ke tui.KeyEvent) { c.count.Set(c.count.Get() - 1) }),
	}
}

templ (c *counterApp) Render() {
	<div class="flex-col gap-1 p-2 items-center justify-center">
		<div class="border-rounded p-1 flex-col items-center justify-center">
			<span class="font-bold text-cyan">Counter</span>
			<hr />
			<span>Count:</span>
			<span class="font-bold text-blue">{fmt.Sprintf("%d", c.count.Get())}</span>
		</div>
		<div class="flex justify-center">
			<span class="font-dim">{"Press +/- to change, q to quit"}</span>
		</div>
	</div>
}
