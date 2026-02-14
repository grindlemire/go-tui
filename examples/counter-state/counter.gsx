package main

import (
	"fmt"
	"time"
	"github.com/grindlemire/go-tui/internal/debug"
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
		tui.OnRune('q', func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnRune('+', func(ke tui.KeyEvent) {
			debug.Log("increment via KeyMap")
			c.count.Set(c.count.Get() + 1)
		}),
		tui.OnRune('-', func(ke tui.KeyEvent) {
			c.count.Set(c.count.Get() - 1)
		}),
	}
}

func (c *counterApp) HandleMouse(me tui.MouseEvent) bool {
	if me.Button == tui.MouseLeft && me.Action == tui.MousePress {
		if c.incrementBtn.El() != nil && c.incrementBtn.El().ContainsPoint(me.X, me.Y) {
			debug.Log("increment via HandleMouse")
			c.count.Set(c.count.Get() + 1)
			return true
		}
		if c.decrementBtn.El() != nil && c.decrementBtn.El().ContainsPoint(me.X, me.Y) {
			c.count.Set(c.count.Get() - 1)
			return true
		}
	}
	return false
}

func (c *counterApp) tick() {
	debug.Log("tick callback called")
	c.count.Set(c.count.Get() + 1)
}

templ (c *counterApp) Render() {
	incrementBtn := c.incrementBtn
	decrementBtn := c.decrementBtn
	<div class="flex-col gap-1 p-2"
	     onTimer={tui.OnTimer(time.Second, c.tick)}>
		<div class="border-rounded p-1 flex-col items-center justify-center">
			<span class="font-bold text-cyan">Reactive Counter</span>
			<hr class="border" />
			<span>{"Count:"}</span>
			<span class="font-bold text-blue">{fmt.Sprintf("%d", c.count.Get())}</span>
		</div>
		<br />
		<div class="flex gap-1 justify-center">
			<button ref={incrementBtn}>{" + "}</button>
			<button ref={decrementBtn}>{" - "}</button>
		</div>
		<div class="flex justify-center">
			<span class="font-dim">{"Press +/- or q to quit"}</span>
		</div>
	</div>
}
