package main

import tui "github.com/grindlemire/go-tui"

// panes keeps prebuilt elements in struct fields and renders them with @expr.
// Any Go expression that yields a *tui.Element (or a tui.Component) works,
// including indexing, so the active pane is picked at render time. Components
// held in a slice or map field are bound to the app through the generated
// BindApp when they are rendered by index or by a loop over the field.
type panes struct {
	active  int
	content []*tui.Element
	items   []*tui.Element
	widgets []tui.Component
	footer  *tui.Element
}

func NewPanes(active int, widgets ...tui.Component) *panes {
	return &panes{
		active:  active,
		widgets: widgets,
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
		for _, w := range p.widgets {
			@w
		}
		@p.footer
	</div>
}
