package main

import tui "github.com/grindlemire/go-tui"

type searchInput struct {
	Active *tui.State[bool]
	Query  *tui.State[string]
}

func SearchInput(active *tui.State[bool], query *tui.State[string]) *searchInput {
	return &searchInput{Active: active, Query: query}
}

func (s *searchInput) KeyMap() tui.KeyMap {
	if !s.Active.Get() {
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
	s.Query.Set(s.Query.Get() + string(ke.Rune))
}

func (s *searchInput) deleteChar(ke tui.KeyEvent) {
	q := s.Query.Get()
	if len(q) > 0 {
		s.Query.Set(q[:len(q)-1])
	}
}

func (s *searchInput) submit(ke tui.KeyEvent) {
	s.Active.Set(false)
}

func (s *searchInput) deactivate(ke tui.KeyEvent) {
	s.Active.Set(false)
	s.Query.Set("")
}

templ (s *searchInput) Render() {
	<div>
		@if s.Active.Get() {
			<div class="border-rounded p-1">
				<span class="text-cyan">Search: </span>
				<span>{s.Query.Get()}</span>
				<span class="font-dim">|</span>
			</div>
		}
	</div>
}
