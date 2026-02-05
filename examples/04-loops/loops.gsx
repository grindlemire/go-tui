package main

import (
	"fmt"
	tui "github.com/grindlemire/go-tui"
)

templ Loops(items []string) {
	selected := tui.NewState(0)
	<div
		class="flex-col p-1 border-rounded gap-1"
		focusable={true}
		onKeyPress={handleKeys(selected, len(items))}>
		<div class="flex justify-between">
			<span class="font-bold text-cyan">{"Loop Rendering"}</span>
			<span class="text-blue font-bold">{fmt.Sprintf("Item %d/%d", selected.Get()+1, len(items))}</span>
		</div>
		<div class="flex gap-1">
			<div class="border-single p-1 flex-col" flexGrow={1.0}>
				<span class="font-bold">{"Simple @for"}</span>
				@for i, item := range items {
					@if i == selected.Get() {
						<span class="text-gradient-cyan-magenta font-bold">{item}</span>
					} @else {
						<span>{item}</span>
					}
				}
			</div>
			<div class="border-single p-1 flex-col" flexGrow={1.0}>
				<span class="font-bold">{"@for with index"}</span>
				@for i, item := range items {
					@if i == selected.Get() {
						<span class="text-gradient-cyan-magenta font-bold">{fmt.Sprintf("%d. %s", i+1, item)}</span>
					} @else {
						<span>{fmt.Sprintf("%d. %s", i+1, item)}</span>
					}
				}
			</div>
			<div class="border-single p-1 flex-col" flexGrow={1.0}>
				<span class="font-bold">{"Selected (reactive)"}</span>
				<span class="text-green font-bold">{items[selected.Get()]}</span>
				<span class="font-dim">{fmt.Sprintf("Index: %d", selected.Get())}</span>
				<span class="font-dim">{fmt.Sprintf("Length: %d chars", len(items[selected.Get()]))}</span>
			</div>
		</div>
		<div class="flex justify-center">
			<span class="font-dim">{"[j/k] navigate  [q] quit"}</span>
		</div>
	</div>
}

func handleKeys(selected *tui.State[int], count int) func(*tui.Element, tui.KeyEvent) bool {
	return func(el *tui.Element, e tui.KeyEvent) bool {
		switch {
		case e.Rune == 'j' || e.Key == tui.KeyDown:
			if selected.Get() < count-1 {
				selected.Set(selected.Get() + 1)
			} else if selected.Get() == count-1 {
				selected.Set(0)
			}
			return true
		case e.Rune == 'k' || e.Key == tui.KeyUp:
			if selected.Get() > 0 {
				selected.Set(selected.Get() - 1)
			} else if selected.Get() == 0 {
				selected.Set(count-1)
			}
			return true
		}
		return false
	}
}
