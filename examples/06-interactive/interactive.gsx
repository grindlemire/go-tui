package main

import (
	"fmt"
	"github.com/grindlemire/go-tui/pkg/layout"
	"github.com/grindlemire/go-tui/pkg/tui"
	"github.com/grindlemire/go-tui/pkg/tui/element"
)

templ Counter() {
	count := tui.NewState(0)
	<div class="flex-col gap-1 p-2 border-rounded">
		<span class="font-bold text-cyan">Interactive Counter</span>
		<hr class="border" />
		<div class="flex gap-2 items-center">
			<span>Count</span>
			<span class="font-bold text-blue">{fmt.Sprintf("%d", count.Get())}</span>
		</div>
		<div class="flex gap-1">
			<button onKeyPress={keyPress(count)} onClick={decrement(count)}>{" - "}</button>
			<button onKeyPress={keyPress(count)} onClick={increment(count)}>{" + "}</button>
			<button onClick={reset(count)}>{" Reset "}</button>
		</div>
		<span class="font-dim">Click buttons or press q to quit</span>
	</div>
}

func increment(count *tui.State[int]) func() {
	return func() {
		count.Set(count.Get() + 1)
	}
}

func decrement(count *tui.State[int]) func() {
	return func() {
		count.Set(count.Get() - 1)
	}
}

func reset(count *tui.State[int]) func() {
	return func() {
		count.Set(0)
	}
}

func keyPress(count *tui.State[int]) func(tui.KeyEvent) {
	return func(e tui.KeyEvent) {
		if e.Rune == '-' {
			count.Set(count.Get() - 1)
		} else if e.Rune == '+' {
			count.Set(count.Get() + 1)
		}
	}
}
