package main

import (
	"fmt"
	tui "github.com/grindlemire/go-tui"
)

type counter struct {
	count        *tui.State[int]
	events       *Events[string]
	decrementBtn *tui.Ref
	incrementBtn *tui.Ref
	resetBtn     *tui.Ref
}

func Counter(events *Events[string]) *counter {
	return &counter{
		count:        tui.NewState(0),
		events:       events,
		decrementBtn: tui.NewRef(),
		incrementBtn: tui.NewRef(),
		resetBtn:     tui.NewRef(),
	}
}

func (c *counter) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnRune('+', func(ke tui.KeyEvent) { c.increment() }),
		tui.OnRune('=', func(ke tui.KeyEvent) { c.increment() }),
		tui.OnRune('-', func(ke tui.KeyEvent) { c.decrement() }),
		tui.OnRune('0', func(ke tui.KeyEvent) { c.reset() }),
	}
}

func (c *counter) HandleMouse(me tui.MouseEvent) bool {
	return tui.HandleClicks(me,
		tui.Click(c.decrementBtn, c.decrement),
		tui.Click(c.incrementBtn, c.increment),
		tui.Click(c.resetBtn, c.reset),
	)
}

func (c *counter) increment() {
	c.count.Set(c.count.Get() + 1)
	c.events.Emit("increment")
}

func (c *counter) decrement() {
	c.count.Set(c.count.Get() - 1)
	c.events.Emit("decrement")
}

func (c *counter) reset() {
	c.count.Set(0)
	c.events.Emit("reset")
}

templ (c *counter) Render() {
	<div class="border-single p-1 flex-col gap-1" flexGrow={1.0}>
		<span class="text-gradient-cyan-blue font-bold">{"Counter"}</span>
		<div class="flex gap-1 items-center">
			<span class="font-dim">Count:</span>
			<span class="text-cyan font-bold">{fmt.Sprintf("%d", c.count.Get())}</span>
		</div>
		<div class="flex gap-1">
			<button ref={c.decrementBtn}>{" - "}</button>
			<button ref={c.incrementBtn}>{" + "}</button>
			<button ref={c.resetBtn}>{" 0 "}</button>
		</div>
		@if c.count.Get() > 0 {
			<span class="text-green font-bold">{"Positive"}</span>
		} @else @if c.count.Get() < 0 {
			<span class="text-red font-bold">{"Negative"}</span>
		} @else {
			<span class="text-blue font-bold">{"Zero"}</span>
		}
		<div flexGrow={1.0}></div>
		<span class="font-dim">{"click btns or +/-/0"}</span>
	</div>
}
