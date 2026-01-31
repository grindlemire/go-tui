package main

import (
	"fmt"
	"time"
	"github.com/grindlemire/go-tui/internal/debug"
	tui "github.com/grindlemire/go-tui"
)

templ CounterUI() {
	count := tui.NewState(0)
	<div class="flex-col gap-1 p-2"
	     onKeyPress={handleKeys(count)}
	     onTimer={tui.OnTimer(time.Second, tick(count))}>
		<div class="border-rounded p-1 flex-col items-center justify-center">
			<span class="font-bold text-cyan">Reactive Counter</span>
			<hr class="border" />
			<span>{"Count:"}</span>
			<span class="font-bold text-blue">{fmt.Sprintf("%d", count.Get())}</span>
		</div>
		<br />
		<div class="flex gap-1 justify-center">
			<button onClick={increment(count)}>{" + "}</button>
			<button onClick={decrement(count)}>{" - "}</button>
		</div>
		<div class="flex justify-center">
			<span class="font-dim">{"Press q to quit"}</span>
		</div>
	</div>
}

func increment(count *tui.State[int]) func(*tui.Element) {
	return func(el *tui.Element) {
		debug.Log("increment callback called")
		count.Set(count.Get() + 1)
	}
}

func decrement(count *tui.State[int]) func(*tui.Element) {
	return func(el *tui.Element) {
		count.Set(count.Get() - 1)
	}
}

func handleKeys(count *tui.State[int]) func(*tui.Element, tui.KeyEvent) {
	return func(el *tui.Element, e tui.KeyEvent) {
		debug.Log("[CounterUI] handleKeys called: %+v", e)
		switch e.Rune {
		case '+':
			count.Set(count.Get() + 1)
		case '-':
			count.Set(count.Get() - 1)
		}
	}
}

func tick(count *tui.State[int]) func() {
	return func() {
		debug.Log("tick callback called")
		count.Set(count.Get() + 1)
	}
}
