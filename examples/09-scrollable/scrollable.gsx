package main

import (
	"fmt"
	tui "github.com/grindlemire/go-tui"
)

templ Scrollable(items []string) {
	<div class="flex-col gap-1 p-1 h-full">
		<span class="font-bold text-cyan">Scrollable Content</span>
		<hr class="border" />
		<div class="flex-col flex-grow overflow-y-scroll border-single p-1"
		     focusable={true}
		     onEvent={handleMouseScroll}
		     onKeyPress={handleScrollKeys}>
			@for i, item := range items {
				<span class={itemStyle(i)}>{fmt.Sprintf("%02d. %s", i+1, item)}</span>
			}
		</div>
		<span class="font-dim w-full gap-1 flex flex-row">
			<span class="font-bold">Use arrow keys or j</span>
			<span class="font-bold">k to scroll</span>
			<span class="font-bold">q to quit</span>
		</span>
	</div>
}

func handleScrollKeys(el *tui.Element, e tui.KeyEvent) {
	switch e.Rune {
	case 'j':
		el.ScrollBy(0, 1)
		return
	case 'k':
		el.ScrollBy(0, -1)
		return
	}
	switch e.Key {
	case tui.KeyDown:
		el.ScrollBy(0, 1)
		return
	case tui.KeyUp:
		el.ScrollBy(0, -1)
		return
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
