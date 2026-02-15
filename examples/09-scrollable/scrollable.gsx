package main

import (
	"fmt"
	tui "github.com/grindlemire/go-tui"
)

type scrollableApp struct {
	items   []string
	scrollY *tui.State[int]
	content *tui.Ref
}

func Scrollable(items []string) *scrollableApp {
	return &scrollableApp{
		items:   items,
		scrollY: tui.NewState(0),
		content: tui.NewRef(),
	}
}

func (s *scrollableApp) scrollBy(delta int) {
	el := s.content.El()
	if el == nil {
		return
	}
	_, maxY := el.MaxScroll()
	newY := s.scrollY.Get() + delta
	if newY < 0 {
		newY = 0
	} else if newY > maxY {
		newY = maxY
	}
	s.scrollY.Set(newY)
}

func (s *scrollableApp) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnRune('q', func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnRune('j', func(ke tui.KeyEvent) { s.scrollBy(1) }),
		tui.OnRune('k', func(ke tui.KeyEvent) { s.scrollBy(-1) }),
		tui.OnKey(tui.KeyDown, func(ke tui.KeyEvent) { s.scrollBy(1) }),
		tui.OnKey(tui.KeyUp, func(ke tui.KeyEvent) { s.scrollBy(-1) }),
	}
}

func (s *scrollableApp) HandleMouse(me tui.MouseEvent) bool {
	switch me.Button {
	case tui.MouseWheelUp:
		s.scrollBy(-1)
		return true
	case tui.MouseWheelDown:
		s.scrollBy(1)
		return true
	}
	return false
}

templ (s *scrollableApp) Render() {
	<div class="flex-col gap-1 p-1 h-full border-rounded">
		<span class="text-gradient-cyan-blue font-bold">Scrollable Content</span>
		<hr class="border" />
		<div
			ref={s.content}
			class="flex-col flex-grow border-single p-1"
			scrollable={tui.ScrollVertical}
			scrollOffset={0, s.scrollY.Get()}>
			@for i, item := range s.items {
				<span class={itemStyle(i)}>{fmt.Sprintf("%02d. %s", i+1, item)}</span>
			}
		</div>
		<span class="font-dim">j/k or arrow keys to scroll, q to quit</span>
	</div>
}
