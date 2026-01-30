package main

import (
	"fmt"
	"github.com/grindlemire/go-tui/pkg/layout"
	"github.com/grindlemire/go-tui/pkg/tui"
	"github.com/grindlemire/go-tui/pkg/tui/element"
)

templ Scrollable(items []string) {
	<div class="flex-col gap-1 p-1 h-full">
		<span class="font-bold text-cyan">Scrollable Content</span>
		<hr class="border" />
		<div #Content class="flex-col flex-grow overflow-y-scroll border-single p-1"
		     focusable={true}
		     onEvent={handleMouseScroll(Content)}
		     onKeyPress={handleScrollKeys(Content)}>
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

func handleScrollKeys(content *element.Element) func(tui.KeyEvent) {
	return func(e tui.KeyEvent) {
		switch e.Rune {
		case 'j':
			content.ScrollBy(0, 1)
			return
		case 'k':
			content.ScrollBy(0, -1)
			return
		}
		switch e.Key {
		case tui.KeyDown:
			content.ScrollBy(0, 1)
			return
		case tui.KeyUp:
			content.ScrollBy(0, -1)
			return
		}
	}
}

func handleMouseScroll(content *element.Element) func(tui.Event) bool {
	return func(e tui.Event) bool {
		if mouse, ok := e.(tui.MouseEvent); ok {
			switch mouse.Button {
			case tui.MouseWheelUp:
				content.ScrollBy(0, -1)
				return true
			case tui.MouseWheelDown:
				content.ScrollBy(0, 1)
				return true
			}
		}
		return false
	}
}
