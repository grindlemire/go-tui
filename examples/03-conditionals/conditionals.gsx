package main

import (
	"fmt"
	tui "github.com/grindlemire/go-tui"
)

templ Conditionals() {
	count := tui.NewState(0)
	<div
		class="flex-col gap-2 p-2 border-rounded"
		scrollable={tui.ScrollVertical}
		focusable={true}
		onEvent={handleMouseScroll}
		onKeyPress={handleKeys(count)}>
		<div class="flex justify-between items-center">
			<span class="font-bold text-cyan">{"Conditional Rendering Demo"}</span>
			<div class="flex gap-1 items-center">
				<span class="font-dim">{"Count:"}</span>
				<span class="font-bold text-blue">{fmt.Sprintf("%d", count.Get())}</span>
			</div>
		</div>
		<hr class="border" />
		<div class="border-single p-1">
			<span class="font-bold">{"Example 1: Reactive Text"}</span>
			<div class="flex-col gap-1 m-1">
				<span class="font-dim">{"Text expressions update live:"}</span>
				<span>{fmt.Sprintf("Count: %d", count.Get())}</span>
				<span>{fmt.Sprintf("Double: %d", count.Get() * 2)}</span>
				<span>{fmt.Sprintf("Is even: %v", count.Get() % 2 == 0)}</span>
			</div>
		</div>

		<div class="border-single p-1">
			<span class="font-bold">{"Example 2: Static Conditionals"}</span>
			<div class="m-1">
				<span class="font-dim">{"These show based on start value:"}</span>
				@if count.Get() == 0 {
					<div class="bg-blue p-1">
						<span>{"Started at zero"}</span>
					</div>
				}
				@if count.Get() > 0 {
					<div class="bg-green p-1">
						<span>{"Started positive"}</span>
					</div>
				}
			</div>
		</div>

		<div class="border-single p-1">
			<span class="font-bold">{"Example 3: If/Else Blocks"}</span>
			<div class="m-1">
				<span class="font-dim">{"Evaluated once at start:"}</span>
				@if count.Get() >= 5 {
					<div class="bg-green p-1">
						<span>{"High count at start"}</span>
					</div>
				} @else {
					<div class="bg-red p-1">
						<span>{"Low count at start"}</span>
					</div>
				}
			</div>
		</div>

		<hr class="border" />
		<div class="flex-col gap-1 items-center">
			<span class="font-dim">{"Press - to decrement, + to increment, r to reset, q to quit"}</span>
			<span class="text-yellow font-dim">{"Note: @if blocks evaluate once at construction"}</span>
			<span class="text-yellow font-dim">{"Text expressions with braces update reactively"}</span>
		</div>
	</div>
}

func handleKeys(count *tui.State[int]) func(*tui.Element, tui.KeyEvent) {
	return func(el *tui.Element, e tui.KeyEvent) {
		switch e.Rune {
		case '+':
			count.Set(count.Get() + 1)
		case '-':
			count.Set(count.Get() - 1)
		case 'r':
			count.Set(0)
		}
	}
}

func handleMouseScroll(el *tui.Element, e tui.Event) bool {
	if mouse, ok := e.(tui.MouseEvent); ok {
		switch mouse.Button {
		case tui.MouseWheelUp:
			el.ScrollBy(0, -1)
			return true
		case tui.MouseWheelDown:
			el.ScrollBy(0, 1)
			return true
		}
	}
	return false
}
