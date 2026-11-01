package main

import (
	"fmt"
	tui "github.com/grindlemire/go-tui"
)

var pages = []string{"Basics", "Lists", "Styles", "Components", "Edge cases"}

const maxLog = 4

// hoverDemo is the root: a tab bar that switches pages, a shared log of hover
// events, and a status line. Each page is its own component in its own file.
type hoverDemo struct {
	inline     bool
	page       *tui.State[string]
	hoveredTab *tui.State[string]
	log        *tui.State[[]string]
	pointer    *tui.State[string]
	moves      *tui.State[int]
	drags      *tui.State[int]
	showModal  *tui.State[bool]
	tabRefs    *tui.RefMap[string]

	basics     *basicsPage
	lists      *listsPage
	styles     *stylesPage
	components *componentsPage
	edges      *edgesPage
}

func HoverDemo(inline bool) *hoverDemo {
	d := &hoverDemo{
		inline:     inline,
		page:       tui.NewState("Basics"),
		hoveredTab: tui.NewState(""),
		log:        tui.NewState([]string{}),
		pointer:    tui.NewState("-"),
		moves:      tui.NewState(0),
		drags:      tui.NewState(0),
		showModal:  tui.NewState(false),
		tabRefs:    tui.NewRefMap[string](),
	}
	d.basics = &basicsPage{d: d}
	d.lists = newListsPage(d)
	d.styles = &stylesPage{d: d}
	d.components = newComponentsPage(d)
	d.edges = newEdgesPage(d)
	return d
}

func (d *hoverDemo) record(event string) {
	entries := append(d.log.Get(), event)
	if len(entries) > maxLog {
		entries = entries[len(entries)-maxLog:]
	}
	d.log.Set(entries)
}

// logged returns a hover callback that only records the event.
func (d *hoverDemo) logged(event string) func() {
	return func() { d.record(event) }
}

// The tabs drive their hover look from state, to show callbacks.
func (d *hoverDemo) enterTab(tab string) func() {
	return func() {
		d.record("enter tab " + tab)
		d.hoveredTab.Set(tab)
	}
}

func (d *hoverDemo) leaveTab(tab string) func() {
	return func() {
		d.record("leave tab " + tab)
		d.hoveredTab.Set("")
	}
}

func (d *hoverDemo) closeModal() {
	d.showModal.Set(false)
}

func (d *hoverDemo) KeyMap() tui.KeyMap {
	km := tui.KeyMap{
		tui.On(tui.Rune('q'), func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.On(tui.Rune('c'), func(ke tui.KeyEvent) { d.log.Set([]string{}) }),
		tui.On(tui.KeyTab, func(ke tui.KeyEvent) { ke.App().FocusNext() }),
		tui.On(tui.KeyTab.Shift(), func(ke tui.KeyEvent) { ke.App().FocusPrev() }),
	}
	for i, p := range pages {
		km = append(km, tui.On(tui.Rune(rune('1'+i)), func(ke tui.KeyEvent) { d.page.Set(p) }))
	}
	if !d.inline {
		km = append(km, tui.On(tui.Rune('m'), func(ke tui.KeyEvent) { d.showModal.Set(true) }))
	}
	return km
}

func (d *hoverDemo) HandleMouse(me tui.MouseEvent) bool {
	d.pointer.Set(fmt.Sprintf("%d,%d", me.X, me.Y))
	switch me.Action {
	case tui.MouseMove:
		d.moves.Set(d.moves.Get() + 1)
	case tui.MouseDrag:
		d.drags.Set(d.drags.Get() + 1)
	}
	if me.Button == tui.MouseLeft && me.Action == tui.MousePress {
		for name, el := range d.tabRefs.All() {
			if el != nil && el.ContainsPoint(me.X, me.Y) {
				d.page.Set(name)
				return true
			}
		}
	}
	return false
}

func (d *hoverDemo) tabClass(tab string) string {
	switch {
	case tab == d.page.Get():
		return "px-1 font-bold bg-cyan text-black"
	case tab == d.hoveredTab.Get():
		return "px-1 underline text-bright-cyan"
	}
	return "px-1 text-white"
}

func (d *hoverDemo) help() string {
	if d.inline {
		return "1-5 or click: pages | Tab focus | c clear log | q quit | no modal inline"
	}
	return "1-5 or click: pages | Tab focus | m modal | c clear log | q quit"
}

templ (d *hoverDemo) Render() {
	<div class="flex-col px-1 border-rounded border-cyan" height={24}>
		<div class="flex gap-2">
			<span class="text-gradient-cyan-magenta font-bold">Hover Demo</span>
			<div class="flex">
				for i, page := range pages {
					<span ref={d.tabRefs} key={page} class={d.tabClass(page)} onMouseEnter={d.enterTab(page)} onMouseLeave={d.leaveTab(page)}>{fmt.Sprintf("%d %s", i+1, page)}</span>
				}
			</div>
		</div>
		<div class="flex mt-1" height={12}>
			if d.page.Get() == "Basics" {
				@d.basics
			} else if d.page.Get() == "Lists" {
				@d.lists
			} else if d.page.Get() == "Styles" {
				@d.styles
			} else if d.page.Get() == "Components" {
				@d.components
			} else {
				@d.edges
			}
		</div>
		<div class="flex-col border-rounded px-1 h-6" borderTitle="Hover events (newest last)">
			for _, entry := range d.log.Get() {
				<span class="font-dim">{entry}</span>
			}
		</div>
		<span class="font-dim">{fmt.Sprintf("pointer %s | %d move events | %d drag events", d.pointer.Get(), d.moves.Get(), d.drags.Get())}</span>
		<span class="font-dim">{d.help()}</span>
		<modal open={d.showModal} class="justify-center items-center">
			<div class="border-rounded p-2 flex-col gap-1 w-40 items-center">
				<span class="font-bold text-yellow">Modal open</span>
				<span class="font-dim">Content behind the backdrop ignores hover and the wheel.</span>
				<button class="px-2 border-single hover:border-green hover:text-green hover:font-bold" onMouseEnter={d.logged("enter modal button")} onMouseLeave={d.logged("leave modal button")} onActivate={d.closeModal}>Close</button>
			</div>
		</modal>
	</div>
}
