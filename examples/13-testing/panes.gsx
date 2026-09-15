package main

import tui "github.com/grindlemire/go-tui"

// panes keeps prebuilt elements in struct fields and renders them with @expr.
// Any Go expression that yields a *tui.Element (or a tui.Component) works,
// including indexing, so the active pane is picked at render time.
type panes struct {
	active  int
	content []*tui.Element
	items   []*tui.Element
	footer  *tui.Element
}

func NewPanes(active int) *panes {
	return &panes{
		active: active,
		content: []*tui.Element{
			tui.New(tui.WithText("Pane A")),
			tui.New(tui.WithText("Pane B")),
		},
		items: []*tui.Element{
			tui.New(tui.WithText("item 1")),
			tui.New(tui.WithText("item 2")),
		},
		footer: tui.New(tui.WithText("footer")),
	}
}

templ (p *panes) Render() {
	<div class="flex-col">
		@p.content[p.active]
		for _, el := range p.items {
			@el
		}
		@p.footer
	</div>
}
