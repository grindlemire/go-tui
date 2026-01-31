package main

import (
	"fmt"
	tui "github.com/grindlemire/go-tui"
)

templ Refs() {
	count := tui.NewState(0)
	counter := tui.NewRef()
	incrementBtn := tui.NewRef()
	decrementBtn := tui.NewRef()
	status := tui.NewRef()
	<div class="flex-col gap-1 p-2 border-rounded">
		<span class="font-bold text-cyan">Named Element References</span>
		<hr class="border" />
		<div ref={counter} class="border-single p-1">
			<span>
				Counter
				{fmt.Sprintf("%d", count.Get())}
			</span>
		</div>
		<div class="flex gap-1 w-full justify-center">
			<button ref={incrementBtn} onClick={handleIncrement(count)} class="border-single text-center p-1 w-10 h-5">{" + "}</button>
			<button ref={decrementBtn} onClick={handleDecrement(count)} class="border-single text-center p-1 w-10 h-5">{" - "}</button>
		</div>
		<div ref={status} class="font-dim">
			<span>Click buttons to update the counter</span>
		</div>
		<span class="font-dim">Press q to quit</span>
	</div>
}

func handleIncrement(count *tui.State[int]) func(*tui.Element) {
	return func(el *tui.Element) {
		count.Set(count.Get() + 1)
	}
}

func handleDecrement(count *tui.State[int]) func(*tui.Element) {
	return func(el *tui.Element) {
		count.Set(count.Get() - 1)
	}
}
