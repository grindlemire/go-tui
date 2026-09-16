package main

import tui "github.com/grindlemire/go-tui"

// panes keeps prebuilt elements in struct fields and renders them with @expr.
type panes struct {
	active  int
	content []*tui.Element
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
		footer: tui.New(tui.WithText("footer")),
	}
}

templ (p *panes) Render() {
	<div class="flex-col">
		@p.content[p.active]
		for _, w := range p.widgets {
			@w
		}
		@p.footer
	</div>
}
