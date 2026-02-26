package main

import tui "github.com/grindlemire/go-tui"

type sidebar struct {
	category *tui.State[string]
	expanded *tui.State[bool]
	selected *tui.State[int]
}

var categories = []string{"Documents", "Images", "Music", "Projects", "Downloads"}

func Sidebar(category *tui.State[string]) *sidebar {
	return &sidebar{
		category: category,
		expanded: tui.NewState(true),
		selected: tui.NewState(0),
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
				s.category.Set(categories[idx])
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
	s.category.Set(categories[s.selected.Get()])
}

func (s *sidebar) moveUp() {
	s.selected.Update(func(v int) int {
		if v <= 0 {
			return len(categories) - 1
		}
		return v - 1
	})
	s.category.Set(categories[s.selected.Get()])
}

func (s *sidebar) sidebarWidth() int {
	if s.expanded.Get() {
		return 22
	}
	return 5
}

templ (s *sidebar) Render() {
	<div class="flex-col border-single shrink-0" width={s.sidebarWidth()}>
		@if s.expanded.Get() {
			<div class="flex-col px-1">
				<span class="text-gradient-cyan-magenta font-bold">Folders</span>
			</div>
			<hr />
			<div class="flex-col px-1 grow">
				@for i, cat := range categories {
					@if i == s.selected.Get() {
						<span class="text-cyan font-bold">{"> " + cat}</span>
					} @else {
						<span class="font-dim">{"  " + cat}</span>
					}
				}
			</div>
			<hr />
			<div class="flex-col px-1">
				<span class="font-dim text-cyan">Ctrl+B hide</span>
			</div>
		} @else {
			<span class="text-cyan font-bold px-1">F</span>
		}
	</div>
}
