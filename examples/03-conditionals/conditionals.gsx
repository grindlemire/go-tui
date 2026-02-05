package main

import (
	"fmt"
	tui "github.com/grindlemire/go-tui"
)

type conditionalsApp struct {
	count *tui.State[int]
}

func Conditionals() *conditionalsApp {
	return &conditionalsApp{
		count: tui.NewState(0),
	}
}

func (c *conditionalsApp) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnRune('q', func(ke tui.KeyEvent) { tui.Stop() }),
		tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { tui.Stop() }),
		tui.OnRune('+', func(ke tui.KeyEvent) { c.count.Set(c.count.Get() + 1) }),
		tui.OnRune('-', func(ke tui.KeyEvent) { c.count.Set(c.count.Get() - 1) }),
		tui.OnRune('r', func(ke tui.KeyEvent) { c.count.Set(0) }),
	}
}

templ (c *conditionalsApp) Render() {
	<div class="flex-col p-1 border-rounded gap-1">
		<div class="flex justify-between">
			<span class="font-bold text-cyan">{"Reactive Conditionals"}</span>
			<span class="text-blue font-bold">{fmt.Sprintf("Count: %d", c.count.Get())}</span>
		</div>
		<div class="flex gap-1">
			<div class="border-single p-1 flex-col" flexGrow={1.0}>
				<span class="font-bold">{"Reactive Text"}</span>
				<span>{fmt.Sprintf("Value:  %d", c.count.Get())}</span>
				<span>{fmt.Sprintf("Double: %d", c.count.Get() * 2)}</span>
				<span>{fmt.Sprintf("Even:   %v", c.count.Get() % 2 == 0)}</span>
			</div>
			<div class="border-single p-1 flex-col" flexGrow={1.0}>
				<span class="font-bold">{"Reactive @if"}</span>
				@if c.count.Get() > 0 {
					<span class="text-green font-bold">{"Positive"}</span>
				}
				@if c.count.Get() == 0 {
					<span class="text-blue font-bold">{"Zero"}</span>
				}
				@if c.count.Get() < 0 {
					<span class="text-red font-bold">{"Negative"}</span>
				}
			</div>
			<div class="border-single p-1 flex-col" flexGrow={1.0}>
				<span class="font-bold">{"Reactive @if/else"}</span>
				@if c.count.Get() >= 5 {
					<span class="text-green font-bold">{"High (5+)"}</span>
				} @else {
					<span class="text-yellow">{"Low (under 5)"}</span>
				}
			</div>
		</div>
		<div class="flex justify-center">
			<span class="font-dim">{"[-] dec  [+] inc  [r] reset  [q] quit"}</span>
		</div>
	</div>
}
