package main

import (
	"fmt"
	tui "github.com/grindlemire/go-tui"
)

type counter struct {
	count *tui.State[int]
}

func NewCounter() *counter {
	return &counter{
		count: tui.NewState(0),
	}
}

func (c *counter) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.On(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.On(tui.Rune('+'), func(ke tui.KeyEvent) {
			c.count.Update(func(v int) int { return v + 1 })
		}),
		tui.On(tui.Rune('-'), func(ke tui.KeyEvent) {
			c.count.Update(func(v int) int { return v - 1 })
		}),
	}
}

templ (c *counter) Render() {
	<div class="flex-col grow items-center justify-center">
		<span class="font-bold">{fmt.Sprintf("Count: %d", c.count.Get())}</span>
		<span class="font-dim">Press + / - to change, Esc to quit</span>
	</div>
}
