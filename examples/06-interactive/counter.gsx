package main

import (
	"fmt"
	"time"
	tui "github.com/grindlemire/go-tui"
)

type counter struct {
	count        *tui.State[int]
	pending      *tui.State[bool]
	events       *tui.Events[string]
	delayCh      chan int
	decrementBtn *tui.Ref
	incrementBtn *tui.Ref
	resetBtn     *tui.Ref
}

var(
	_ tui.Component = (*counter)(nil)
	_ tui.WatcherProvider = (*counter)(nil)
	_ tui.MouseListener = (*counter)(nil)
	_ tui.KeyListener = (*counter)(nil)
)

func Counter(events *tui.Events[string]) *counter {
	return &counter{
		count:        tui.NewState(0),
		pending:      tui.NewState(false),
		events:       events,
		delayCh:      make(chan int),
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
		tui.OnRune('d', func(ke tui.KeyEvent) { c.delayedIncrement() }),
	}
}

func (c *counter) Watchers() []tui.Watcher {
	return []tui.Watcher{
		tui.Watch(c.delayCh, func(delta int) {
			c.count.Set(c.count.Get() + delta)
			c.pending.Set(false)
			c.events.Emit("delayed +1")
		}),
	}
}

func (c *counter) delayedIncrement() {
	if c.pending.Get() {
		return // Already pending
	}
	c.pending.Set(true)
	c.events.Emit("delay started")
	go func() {
		time.Sleep(1 * time.Second)
		c.delayCh <- 1
	}()
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
	<div class="border-single p-1 flex-col gap-1 grow justify-center w-1/2">
		<span class="text-gradient-cyan-blue font-bold text-center">{"Counter"}</span>
		<div class="flex gap-1 items-center justify-center">
			<span class="font-dim">Count:</span>
			<span class="text-cyan font-bold">{fmt.Sprintf("%d", c.count.Get())}</span>
		</div>
		<div class="flex gap-1 justify-center">
			<button ref={c.decrementBtn}>{" - "}</button>
			<button ref={c.incrementBtn}>{" + "}</button>
			<button ref={c.resetBtn}>{" 0 "}</button>
		</div>
		@if c.pending.Get() {
			<span class="text-yellow font-bold text-center">{"Pending..."}</span>
		} @else @if c.count.Get() > 0 {
			<span class="text-green font-bold text-center">{"Positive"}</span>
		} @else @if c.count.Get() < 0 {
			<span class="text-red font-bold text-center">{"Negative"}</span>
		} @else {
			<span class="text-blue font-bold text-center">{"Zero"}</span>
		}
	</div>
}
