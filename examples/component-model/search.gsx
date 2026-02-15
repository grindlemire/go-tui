package main

import tui "github.com/grindlemire/go-tui"

type searchInput struct {
	active *tui.State[bool]
	query  *tui.State[string]
}

func SearchInput(active *tui.State[bool], query *tui.State[string]) *searchInput {
	return &searchInput{active: active, query: query}
}

func (s *searchInput) KeyMap() tui.KeyMap {
	if !s.active.Get() {
		return nil
	}
	return tui.KeyMap{
		tui.OnRunesStop(s.appendChar),
		tui.OnKeyStop(tui.KeyBackspace, s.deleteChar),
		tui.OnKeyStop(tui.KeyEnter, s.submit),
		tui.OnKeyStop(tui.KeyEscape, s.deactivate),
	}
}

func (s *searchInput) appendChar(ke tui.KeyEvent) {
	s.query.Set(s.query.Get() + string(ke.Rune))
}

func (s *searchInput) deleteChar(ke tui.KeyEvent) {
	q := s.query.Get()
	if len(q) > 0 {
		s.query.Set(q[:len(q)-1])
	}
}

func (s *searchInput) submit(ke tui.KeyEvent) {
	s.active.Set(false)
}

func (s *searchInput) deactivate(ke tui.KeyEvent) {
	s.active.Set(false)
	s.query.Set("")
}

templ (s *searchInput) Render() {
	<div>
		@if s.active.Get() {
			<div class="border-rounded p-1">
				<span class="text-cyan">Search: </span>
				<span>{s.query.Get()}</span>
				<span class="font-dim">|</span>
			</div>
		}
	</div>
}
