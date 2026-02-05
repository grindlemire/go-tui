package main

import (
	"fmt"
	tui "github.com/grindlemire/go-tui"
)

type scrollableApp struct {
	items []string
}

func Scrollable(items []string) *scrollableApp {
	return &scrollableApp{
		items: items,
	}
}

func (s *scrollableApp) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnRune('q', func(ke tui.KeyEvent) { tui.Stop() }),
		tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { tui.Stop() }),
	}
}

templ (s *scrollableApp) Render() {
	<div class="flex-col gap-1 p-1 h-full">
		<span class="font-bold text-cyan">Scrollable Content</span>
		<hr class="border" />
		<div class="flex-col flex-grow overflow-y-scroll border-single p-1"
		     onEvent={handleMouseScroll}
		     onKeyPress={handleScrollKeys}
		     focusable={true}>
			@for i, item := range s.items {
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

func handleScrollKeys(el *tui.Element, e tui.KeyEvent) bool {
	switch e.Rune {
	case 'j':
		el.ScrollBy(0, 1)
		return true
	case 'k':
		el.ScrollBy(0, -1)
		return true
	}
	switch e.Key {
	case tui.KeyDown:
		el.ScrollBy(0, 1)
		return true
	case tui.KeyUp:
		el.ScrollBy(0, -1)
		return true
	}
	return false
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
