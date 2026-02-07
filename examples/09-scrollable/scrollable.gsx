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
	<div class="flex-col gap-1 p-1 h-full border-rounded">
		<span class="text-gradient-cyan-blue font-bold">{"Scrollable Content"}</span>
		<hr class="border" />
		<div
			class="flex-col flex-grow border-single p-1"
			scrollable={tui.ScrollVertical}
			focusable={true}>
			@for i, item := range s.items {
				<span class={itemStyle(i)}>{fmt.Sprintf("%02d. %s", i+1, item)}</span>
			}
		</div>
		<span class="font-dim">{"Arrow keys/Page Up/Down to scroll, q to quit"}</span>
	</div>
}
