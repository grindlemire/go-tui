package main

import (
	"fmt"
	tui "github.com/grindlemire/go-tui"
)

type loopsApp struct {
	items    []string
	selected *tui.State[int]
}

func Loops(items []string) *loopsApp {
	return &loopsApp{
		items:    items,
		selected: tui.NewState(0),
	}
}

func (l *loopsApp) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnRune('q', func(ke tui.KeyEvent) { tui.Stop() }),
		tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { tui.Stop() }),
		tui.OnRune('j', func(ke tui.KeyEvent) { l.next() }),
		tui.OnKey(tui.KeyDown, func(ke tui.KeyEvent) { l.next() }),
		tui.OnRune('k', func(ke tui.KeyEvent) { l.prev() }),
		tui.OnKey(tui.KeyUp, func(ke tui.KeyEvent) { l.prev() }),
	}
}

func (l *loopsApp) next() {
	if l.selected.Get() < len(l.items)-1 {
		l.selected.Set(l.selected.Get() + 1)
	} else {
		l.selected.Set(0)
	}
}

func (l *loopsApp) prev() {
	if l.selected.Get() > 0 {
		l.selected.Set(l.selected.Get() - 1)
	} else {
		l.selected.Set(len(l.items) - 1)
	}
}

templ (l *loopsApp) Render() {
	<div class="flex-col p-1 border-rounded gap-1">
		<div class="flex justify-between">
			<span class="font-bold text-cyan">{"Loop Rendering"}</span>
			<span class="text-blue font-bold">{fmt.Sprintf("Item %d/%d", l.selected.Get()+1, len(l.items))}</span>
		</div>
		<div class="flex gap-1">
			<div class="border-single p-1 flex-col" flexGrow={1.0}>
				<span class="font-bold">{"Simple @for"}</span>
				@for i, item := range l.items {
					@if i == l.selected.Get() {
						<span class="text-gradient-cyan-magenta font-bold">{item}</span>
					} @else {
						<span>{item}</span>
					}
				}
			</div>
			<div class="border-single p-1 flex-col" flexGrow={1.0}>
				<span class="font-bold">{"@for with index"}</span>
				@for i, item := range l.items {
					@if i == l.selected.Get() {
						<span class="text-gradient-cyan-magenta font-bold">{fmt.Sprintf("%d. %s", i+1, item)}</span>
					} @else {
						<span>{fmt.Sprintf("%d. %s", i+1, item)}</span>
					}
				}
			</div>
			<div class="border-single p-1 flex-col" flexGrow={1.0}>
				<span class="font-bold">{"Selected (reactive)"}</span>
				<span class="text-green font-bold">{l.items[l.selected.Get()]}</span>
				<span class="font-dim">{fmt.Sprintf("Index: %d", l.selected.Get())}</span>
				<span class="font-dim">{fmt.Sprintf("Length: %d chars", len(l.items[l.selected.Get()]))}</span>
			</div>
		</div>
		<div class="flex justify-center">
			<span class="font-dim">{"[j/k] navigate  [q] quit"}</span>
		</div>
	</div>
}
