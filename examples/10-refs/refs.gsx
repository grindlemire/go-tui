package main

import (
	"fmt"
	"github.com/grindlemire/go-tui/pkg/layout"
	"github.com/grindlemire/go-tui/pkg/tui"
	"github.com/grindlemire/go-tui/pkg/tui/element"
)

templ Refs() {
	count := tui.NewState(0)
	<div class="flex-col gap-1 p-2 border-rounded">
		<span class="font-bold text-cyan">Named Element References</span>
		<hr class="border" />
		<div #Counter class="border-single p-1">
			<span>
				Counter
				{fmt.Sprintf("%d", count.Get())}
			</span>
		</div>
		<div class="flex gap-1 w-full justify-center">
			<button #IncrementBtn onClick={handleIncrement(count)} class="border-single text-center p-1 w-10 h-5">{" + "}</button>
			<button #DecrementBtn onClick={handleDecrement(count)} class="border-single text-center p-1 w-10 h-5">{" - "}</button>
		</div>
		<div #Status class="font-dim">
			<span>Click buttons to update the counter</span>
		</div>
		<span class="font-dim">Press q to quit</span>
	</div>
}

func handleIncrement(count *tui.State[int]) func() {
	return func() {
		count.Set(count.Get() + 1)
	}
}

func handleDecrement(count *tui.State[int]) func() {
	return func() {
		count.Set(count.Get() - 1)
	}
}
