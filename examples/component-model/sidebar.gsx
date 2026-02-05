package main

import tui "github.com/grindlemire/go-tui"

type sidebar struct {
	Query    *tui.State[string]
	expanded *tui.State[bool]
}

func Sidebar(query *tui.State[string]) *sidebar {
	return &sidebar{
		Query:    query,
		expanded: tui.NewState(true),
	}
}

func (s *sidebar) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnKey(tui.KeyCtrlB, s.toggle),
	}
}

func (s *sidebar) toggle(ke tui.KeyEvent) {
	s.expanded.Set(!s.expanded.Get())
}

templ (s *sidebar) Render() {
	<div>
		@if s.expanded.Get() {
			<div class="flex-col border-single p-1" width={30}>
				<span class="font-bold">Sidebar</span>
				@if s.Query.Get() != "" {
					<span>Filtering: {s.Query.Get()}</span>
				}
				<span class="font-dim">Ctrl+B to toggle</span>
			</div>
		}
	</div>
}
