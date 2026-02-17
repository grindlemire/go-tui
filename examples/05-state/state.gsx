package main

import (
	"fmt"
	tui "github.com/grindlemire/go-tui"
)

// panel is a reusable struct component that accepts children via {children...}.
type panel struct {
	title    string
	children []*tui.Element
}

func NewPanel(title string, children []*tui.Element) *panel {
	return &panel{title: title, children: children}
}

templ (p *panel) Render() {
	<div class="flex-col border-rounded p-1 gap-1">
		<span class="text-gradient-cyan-magenta font-bold">{p.title}</span>
		{children...}
	</div>
}

type stateApp struct {
	count    *tui.State[int]
	selected *tui.State[int]
	items    []string
}

func State() *stateApp {
	return &stateApp{
		count:    tui.NewState(0),
		selected: tui.NewState(0),
		items:    []string{"Rust", "Go", "TypeScript", "Python", "Zig"},
	}
}

func (s *stateApp) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnRune('q', func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnRune('+', func(ke tui.KeyEvent) { s.count.Update(func(v int) int { return v + 1 }) }),
		tui.OnRune('=', func(ke tui.KeyEvent) { s.count.Update(func(v int) int { return v + 1 }) }),
		tui.OnRune('-', func(ke tui.KeyEvent) { s.count.Update(func(v int) int { return v - 1 }) }),
		tui.OnRune('r', func(ke tui.KeyEvent) { s.count.Set(0) }),
		tui.OnRune('j', func(ke tui.KeyEvent) { s.selectNext() }),
		tui.OnRune('k', func(ke tui.KeyEvent) { s.selectPrev() }),
		tui.OnKey(tui.KeyDown, func(ke tui.KeyEvent) { s.selectNext() }),
		tui.OnKey(tui.KeyUp, func(ke tui.KeyEvent) { s.selectPrev() }),
	}
}

func (s *stateApp) selectNext() {
	n := len(s.items)
	s.selected.Update(func(v int) int {
		if v >= n-1 {
			return 0
		}
		return v + 1
	})
}

func (s *stateApp) selectPrev() {
	n := len(s.items)
	s.selected.Update(func(v int) int {
		if v <= 0 {
			return n - 1
		}
		return v - 1
	})
}

func isEven(n int) bool {
	return n/2*2 == n
}

func rangeLabel(count int) string {
	if count > 20 {
		return "high (> 20)"
	}
	if count > 0 {
		return "medium (1-20)"
	}
	if count == 0 {
		return "zero"
	}
	return "negative"
}

templ (s *stateApp) Render() {
	@let spanCount = <span class="text-cyan font-bold">{fmt.Sprintf("%d", s.count.Get())}</span>
	<div class="flex-col p-1 border-rounded border-cyan">
		<span class="text-gradient-cyan-magenta font-bold">State and Control Flow</span>

		// Top row: Counter + Status
		<div class="flex">
			// Counter panel
			<div class="flex-col border-rounded p-1 gap-1 items-center justify-center" flexGrow={1.0}>
				<span class="text-gradient-cyan-magenta font-bold">Counter</span>
				<br />
				{spanCount}
				<div class="flex gap-1 justify-center">
					<span class="text-cyan px-1 font-bold">+</span>
					<span class="text-cyan px-1 font-bold">-</span>
					<span class="text-cyan px-1 font-bold">r</span>
				</div>
			</div>

			// Status panel with @if/@else
			<div class="flex-col border-rounded p-1 gap-1" flexGrow={2.0}>
				<span class="text-gradient-cyan-magenta font-bold">Status</span>
				<div class="flex gap-1">
					<span class="font-dim">Sign:</span>
					@if s.count.Get() > 0 {
						<span class="text-green font-bold">Positive</span>
					} @else @if s.count.Get() < 0 {
						<span class="text-red font-bold">Negative</span>
					} @else {
						<span class="text-blue font-bold">Zero</span>
					}
				</div>
				<div class="flex gap-1">
					<span class="font-dim">Parity:</span>
					@if isEven(s.count.Get()) {
						<span class="text-cyan">Even</span>
					} @else {
						<span class="text-magenta">Odd</span>
					}
				</div>
				<div class="flex gap-1">
					<span class="font-dim">Range:</span>
					<span class="text-yellow">{rangeLabel(s.count.Get())}</span>
				</div>
			</div>
		</div>

		// Items list with @for — uses panel struct component with {children...}
		@NewPanel("Items") {
			@for i, item := range s.items {
				@if i == s.selected.Get() {
					<span class="text-gradient-cyan-magenta font-bold">{fmt.Sprintf("  > %s", item)}</span>
				} @else {
					<span class="font-dim">{fmt.Sprintf("    %s", item)}</span>
				}
			}
		}

		// Key hints
		<div class="flex justify-center">
			<span class="font-dim">+/-count|j/k navigate|r reset|q quit</span>
		</div>
	</div>
}
