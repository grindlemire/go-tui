package main

import tui "github.com/grindlemire/go-tui"

type sidebar struct {
	categoryBus *tui.Events[string]
	expanded    *tui.State[bool]
	selected    *tui.State[int]
}

var categories = []string{"Documents", "Images", "Music", "Projects", "Downloads"}

func Sidebar() *sidebar {
	return &sidebar{
		categoryBus: tui.NewEvents[string](categoryTopic),
		expanded:    tui.NewState(true),
		selected:    tui.NewState(0),
	}
}

func (s *sidebar) KeyMap() tui.KeyMap {
	km := tui.KeyMap{
		tui.OnKey(tui.KeyCtrlB, func(ke tui.KeyEvent) {
			s.expanded.Set(!s.expanded.Get())
		}),
	}
	if s.expanded.Get() {
		km = append(km, tui.OnRune('j', func(ke tui.KeyEvent) { s.moveDown() }))
		km = append(km, tui.OnRune('k', func(ke tui.KeyEvent) { s.moveUp() }))
		km = append(km, tui.OnKey(tui.KeyDown, func(ke tui.KeyEvent) { s.moveDown() }))
		km = append(km, tui.OnKey(tui.KeyUp, func(ke tui.KeyEvent) { s.moveUp() }))
		km = append(km, tui.OnKey(tui.KeyEnter, func(ke tui.KeyEvent) {
			idx := s.selected.Get()
			if idx >= 0 && idx < len(categories) {
				s.categoryBus.Emit(categories[idx])
			}
		}))
	}
	return km
}

func (s *sidebar) moveDown() {
	s.selected.Update(func(v int) int {
		if v >= len(categories)-1 {
			return 0
		}
		return v + 1
	})
	s.categoryBus.Emit(categories[s.selected.Get()])
}

func (s *sidebar) moveUp() {
	s.selected.Update(func(v int) int {
		if v <= 0 {
			return len(categories) - 1
		}
		return v - 1
	})
	s.categoryBus.Emit(categories[s.selected.Get()])
}

templ (s *sidebar) Render() {
	<div>
		@if s.expanded.Get() {
			<div class="flex-col border-single p-1 gap-1" width={20}>
				<span class="text-gradient-cyan-magenta font-bold">Folders</span>
				<hr />
				@for i, cat := range categories {
					@if i == s.selected.Get() {
						<span class="text-cyan font-bold">{"> " + cat}</span>
					} @else {
						<span class="font-dim">{"  " + cat}</span>
					}
				}
				<hr />
				<span class="font-dim">Ctrl+B hide</span>
			</div>
		} @else {
			<div class="flex-col border-single p-1" width={4}>
				<span class="text-cyan font-bold">F</span>
			</div>
		}
	</div>
}
