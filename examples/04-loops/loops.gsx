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
		tui.OnRune('q', func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnRune('j', func(ke tui.KeyEvent) { l.next() }),
		tui.OnKey(tui.KeyDown, func(ke tui.KeyEvent) { l.next() }),
		tui.OnRune('k', func(ke tui.KeyEvent) { l.prev() }),
		tui.OnKey(tui.KeyUp, func(ke tui.KeyEvent) { l.prev() }),
	}
}

func (l *loopsApp) next() {
	itemCount := len(l.items)
	if itemCount == 0 {
		return
	}
	l.selected.Update(func(current int) int {
		if current < 0 || current >= itemCount-1 {
			return 0
		}
		return current + 1
	})
}

func (l *loopsApp) prev() {
	itemCount := len(l.items)
	if itemCount == 0 {
		return
	}
	l.selected.Update(func(current int) int {
		if current <= 0 || current >= itemCount {
			return itemCount - 1
		}
		return current - 1
	})
}

func (l *loopsApp) selectedIndex() int {
	itemCount := len(l.items)
	if itemCount == 0 {
		return -1
	}
	index := l.selected.Get()
	if index < 0 {
		return 0
	}
	if index >= itemCount {
		return itemCount - 1
	}
	return index
}

func (l *loopsApp) selectedItem() string {
	index := l.selectedIndex()
	if index < 0 {
		return ""
	}
	return l.items[index]
}

templ (l *loopsApp) Render() {
	<div class="flex-col p-1 border-rounded gap-1">
		<div class="flex justify-between">
			<span class="font-bold text-cyan">{"Loop Rendering"}</span>
			<span class="text-blue font-bold">{fmt.Sprintf("Item %d/%d", l.selectedIndex()+1, len(l.items))}</span>
		</div>
		<div class="flex gap-1">
			<div class="border-single p-1 flex-col" flexGrow={1.0}>
				<span class="font-bold">{"Simple @for"}</span>
				@for i, item := range l.items {
					@if i == l.selectedIndex() {
						<span class="text-gradient-cyan-magenta font-bold">{item}</span>
					} @else {
						<span>{item}</span>
					}
				}
			</div>
			<div class="border-single p-1 flex-col" flexGrow={1.0}>
				<span class="font-bold">{"@for with index"}</span>
				@for i, item := range l.items {
					@if i == l.selectedIndex() {
						<span class="text-gradient-cyan-magenta font-bold">{fmt.Sprintf("%d. %s", i+1, item)}</span>
					} @else {
						<span>{fmt.Sprintf("%d. %s", i+1, item)}</span>
					}
				}
			</div>
			<div class="border-single p-1 flex-col" flexGrow={1.0}>
				<span class="font-bold">{"Selected (reactive)"}</span>
				@if len(l.items) == 0 {
					<span class="text-yellow font-bold">{"(no items)"}</span>
					<span class="font-dim">{"Index: n/a"}</span>
					<span class="font-dim">{"Length: n/a"}</span>
				} @else {
					<span class="text-green font-bold">{l.selectedItem()}</span>
					<span class="font-dim">{fmt.Sprintf("Index: %d", l.selectedIndex())}</span>
					<span class="font-dim">{fmt.Sprintf("Length: %d chars", len(l.selectedItem()))}</span>
				}
			</div>
		</div>
		<div class="flex justify-center">
			<span class="font-dim">{"[j/k] navigate  [q] quit"}</span>
		</div>
	</div>
}
