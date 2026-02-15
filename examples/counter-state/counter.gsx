package main

import (
	"fmt"
	tui "github.com/grindlemire/go-tui"
)

type counterApp struct {
	count        *tui.State[int]
	incrementBtn *tui.Ref
	decrementBtn *tui.Ref
}

func Counter() *counterApp {
	return &counterApp{
		count:        tui.NewState(0),
		incrementBtn: tui.NewRef(),
		decrementBtn: tui.NewRef(),
	}
}

func (c *counterApp) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnKey(tui.KeyCtrlC, func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnRune('q', func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnRune('+', func(ke tui.KeyEvent) { c.increment() }),
		tui.OnRune('-', func(ke tui.KeyEvent) { c.decrement() }),
	}
}

func (c *counterApp) HandleMouse(me tui.MouseEvent) bool {
	return tui.HandleClicks(me,
		tui.Click(c.incrementBtn, c.increment),
		tui.Click(c.decrementBtn, c.decrement),
	)
}

func (c *counterApp) increment() {
	c.count.Set(c.count.Get() + 1)
}

func (c *counterApp) decrement() {
	c.count.Set(c.count.Get() - 1)
}

templ (c *counterApp) Render() {
	<div class="flex-col gap-1 p-2">
		<div class="border-rounded p-1 flex-col items-center justify-center">
			<span class="font-bold text-cyan">Reactive Counter</span>
			<hr />
			<span>Count:</span>
			<span class="font-bold text-blue">{fmt.Sprintf("%d", c.count.Get())}</span>
		</div>
		<div class="flex gap-1 justify-center">
			<button ref={c.incrementBtn}>{" + "}</button>
			<button ref={c.decrementBtn}>{" - "}</button>
		</div>
		<div class="flex justify-center">
			<span class="font-dim">{"Press +/- or q to quit"}</span>
		</div>
	</div>
}
