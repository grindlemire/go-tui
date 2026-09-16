// Element expressions: @expr renders any *tui.Element or tui.Component value.
package testdata

import tui "github.com/grindlemire/go-tui"

type tabs struct {
	active  int
	content []*tui.Element
	items   []*tui.Element
	footer  *tui.Element
}

func Tabs(active int, content []*tui.Element) *tabs {
	return &tabs{active: active, content: content}
}

templ (t *tabs) Render() {
	current := @t.content[t.active]
	<div class="flex-col">
		{current}
		for _, el := range t.items {
			@el
		}
		if t.footer != nil {
			@t.footer
		}
	</div>
}
