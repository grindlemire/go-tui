package main

import (
	"fmt"
	tui "github.com/grindlemire/go-tui"
)

templ Conditionals() {
	count := tui.NewState(0)
	<div
		class="flex-col p-1 border-rounded gap-1"
		focusable={true}
		onKeyPress={handleKeys(count)}>
		<div class="flex justify-between">
			<span class="font-bold text-cyan">{"Reactive Conditionals"}</span>
			<span class="text-blue font-bold">{fmt.Sprintf("Count: %d", count.Get())}</span>
		</div>
		<div class="flex gap-1">
			<div class="border-single p-1 flex-col" flexGrow={1.0}>
				<span class="font-bold">{"Reactive Text"}</span>
				<span>{fmt.Sprintf("Value:  %d", count.Get())}</span>
				<span>{fmt.Sprintf("Double: %d", count.Get() * 2)}</span>
				<span>{fmt.Sprintf("Even:   %v", count.Get() % 2 == 0)}</span>
			</div>
			<div class="border-single p-1 flex-col" flexGrow={1.0}>
				<span class="font-bold">{"Reactive @if"}</span>
				@if count.Get() > 0 {
					<span class="text-green font-bold">{"Positive"}</span>
				}
				@if count.Get() == 0 {
					<span class="text-blue font-bold">{"Zero"}</span>
				}
				@if count.Get() < 0 {
					<span class="text-red font-bold">{"Negative"}</span>
				}
			</div>
			<div class="border-single p-1 flex-col" flexGrow={1.0}>
				<span class="font-bold">{"Reactive @if/else"}</span>
				@if count.Get() >= 5 {
					<span class="text-green font-bold">{"High (5+)"}</span>
				} @else {
					<span class="text-yellow">{"Low (under 5)"}</span>
				}
			</div>
		</div>
		<div class="flex justify-center">
			<span class="font-dim">{"[-] dec  [+] inc  [r] reset  [q] quit"}</span>
		</div>
	</div>
}

func handleKeys(count *tui.State[int]) func(*tui.Element, tui.KeyEvent) bool {
	return func(el *tui.Element, e tui.KeyEvent) bool {
		switch e.Rune {
		case '+':
			count.Set(count.Get() + 1)
			return true
		case '-':
			count.Set(count.Get() - 1)
			return true
		case 'r':
			count.Set(0)
			return true
		}
		return false
	}
}
