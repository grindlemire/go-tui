package main

import (
	"fmt"
	tui "github.com/grindlemire/go-tui"
)

// Helper function: formats an item label
func itemLabel(index int, name string, selected bool) string {
	prefix := "  "
	if selected {
		prefix = "> "
	}
	return fmt.Sprintf("%s%d. %s", prefix, index+1, name)
}

// Helper function: returns a style class based on selection
func itemClass(selected bool) string {
	if selected {
		return "text-cyan font-bold"
	}
	return ""
}

// Pure component with children slot
templ Panel(title string) {
	<div class="border-rounded p-1 flex-col gap-1" width={32}>
		<span class="font-bold text-gradient-cyan-magenta">{title}</span>
		<hr />
		{children...}
	</div>
}

// Struct component with state and key handling
type listApp struct {
	items    []string
	selected *tui.State[int]
}

func ListApp() *listApp {
	return &listApp{
		items:    []string{"Alpha", "Bravo", "Charlie", "Delta"},
		selected: tui.NewState(0),
	}
}

func (l *listApp) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnKey(tui.KeyUp, func(ke tui.KeyEvent) {
			l.selected.Update(func(v int) int {
				if v > 0 {
					return v - 1
				}
				return v
			})
		}),
		tui.OnKey(tui.KeyDown, func(ke tui.KeyEvent) {
			l.selected.Update(func(v int) int {
				if v < len(l.items)-1 {
					return v + 1
				}
				return v
			})
		}),
	}
}

templ (l *listApp) Render() {
	<div class="flex-col items-center justify-center h-full">
		@Panel("Select an Item") {
			@for i, item := range l.items {
				<span class={itemClass(i == l.selected.Get())}>
					{itemLabel(i, item, i == l.selected.Get())}
				</span>
			}
			<br />
			@if l.selected.Get() >= 0 {
				<span class="font-dim">{fmt.Sprintf("Selected: %s", l.items[l.selected.Get()])}</span>
			}
		}
	</div>
}
