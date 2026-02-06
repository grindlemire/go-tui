package main

import (
	"fmt"
	tui "github.com/grindlemire/go-tui"
)

type scrollableApp struct {
	items   []string
	content *tui.Ref
}

func Scrollable(items []string) *scrollableApp {
	return &scrollableApp{
		items:   items,
		content: tui.NewRef(),
	}
}

func (s *scrollableApp) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnRune('q', func(ke tui.KeyEvent) { tui.Stop() }),
		tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { tui.Stop() }),
		// Custom j/k scrolling
		tui.OnRune('j', func(ke tui.KeyEvent) {
			if s.content.El() != nil {
				s.content.El().ScrollBy(0, 1)
			}
		}),
		tui.OnRune('k', func(ke tui.KeyEvent) {
			if s.content.El() != nil {
				s.content.El().ScrollBy(0, -1)
			}
		}),
	}
}

templ (s *scrollableApp) Render() {
	content := s.content
	<div class="flex-col gap-1 p-1 h-full">
		<span class="font-bold text-cyan">Scrollable Content</span>
		<hr class="border" />
		<div ref={content}
		     class="flex-col flex-grow overflow-y-scroll border-single p-1"
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
