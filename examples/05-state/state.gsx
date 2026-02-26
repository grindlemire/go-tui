package main

import (
	"fmt"

	tui "github.com/grindlemire/go-tui"
)

type demoApp struct {
	count    *tui.State[int]
	selected *tui.State[int]
	items    []string
}

func Demo() *demoApp {
	return &demoApp{
		count:    tui.NewState(0),
		selected: tui.NewState(0),
		items:    []string{"Rust", "Go", "TypeScript", "Python", "Zig"},
	}
}

func (d *demoApp) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnRune('+', func(ke tui.KeyEvent) {
			d.count.Update(func(v int) int { return v + 1 })
		}),
		tui.OnRune('-', func(ke tui.KeyEvent) {
			d.count.Update(func(v int) int { return v - 1 })
		}),
		tui.OnRune('r', func(ke tui.KeyEvent) {
			ke.App().Batch(func() {
				d.count.Set(0)
				d.selected.Set(0)
			})
		}),
		tui.OnRune('j', func(ke tui.KeyEvent) { d.selectNext() }),
		tui.OnRune('k', func(ke tui.KeyEvent) { d.selectPrev() }),
		tui.OnKey(tui.KeyDown, func(ke tui.KeyEvent) { d.selectNext() }),
		tui.OnKey(tui.KeyUp, func(ke tui.KeyEvent) { d.selectPrev() }),
	}
}

func (d *demoApp) selectNext() {
	d.selected.Update(func(v int) int {
		if v >= len(d.items)-1 {
			return 0
		}
		return v + 1
	})
}

func (d *demoApp) selectPrev() {
	d.selected.Update(func(v int) int {
		if v <= 0 {
			return len(d.items) - 1
		}
		return v - 1
	})
}

func signLabel(n int) string {
	if n > 0 {
		return "Positive"
	}
	if n < 0 {
		return "Negative"
	}
	return "Zero"
}

func signClass(n int) string {
	if n > 0 {
		return "text-green font-bold"
	}
	if n < 0 {
		return "text-red font-bold"
	}
	return "text-blue font-bold"
}

templ (d *demoApp) Render() {
	<div class="flex-col p-1 border-rounded border-cyan">
		<span class="text-gradient-cyan-magenta font-bold">State Demo</span>

		<div class="flex">
			<div class="flex-col border-rounded p-1 gap-1 items-center justify-center" flexGrow={1.0}>
				<span class="font-bold">Counter</span>
				<span class="text-cyan font-bold">{fmt.Sprintf("%d", d.count.Get())}</span>
				<div class="flex gap-1 justify-center">
					<span class="font-dim">+/-  r:reset</span>
				</div>
			</div>

			<div class="flex-col border-rounded p-1 gap-1" flexGrow={2.0}>
				<span class="font-bold">Status</span>
				<div class="flex gap-1">
					<span class="font-dim">Sign:</span>
					<span class={signClass(d.count.Get())}>{signLabel(d.count.Get())}</span>
				</div>
				<div class="flex gap-1">
					<span class="font-dim">Parity:</span>
					@if d.count.Get()%2 == 0 {
						<span class="text-cyan">Even</span>
					} @else {
						<span class="text-magenta">Odd</span>
					}
				</div>
			</div>
		</div>

		<div class="flex-col border-rounded p-1 gap-1">
			<span class="font-bold">Languages</span>
			@for i, item := range d.items {
				@if i == d.selected.Get() {
					<span class="text-cyan font-bold">{fmt.Sprintf("  > %s", item)}</span>
				} @else {
					<span class="font-dim">{fmt.Sprintf("    %s", item)}</span>
				}
			}
		</div>

		<div class="flex justify-center">
			<span class="font-dim">+/- count | j/k navigate | r reset | esc quit</span>
		</div>
	</div>
}
