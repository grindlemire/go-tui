package main

import (
	"fmt"
	tui "github.com/grindlemire/go-tui"
)

// edgesPage stresses hover with a handler that moves its own element out
// from under the pointer.
type edgesPage struct {
	d     *hoverDemo
	right *tui.State[bool]
	dodge *tui.State[int]
}

func newEdgesPage(d *hoverDemo) *edgesPage {
	return &edgesPage{
		d:     d,
		right: tui.NewState(false),
		dodge: tui.NewState(0),
	}
}

// moveAway runs when the dodge button is entered and moves it to the other
// side, so it gets its leave right after the render with no pointer movement.
func (p *edgesPage) moveAway() {
	p.d.record("enter dodge button (moves away)")
	p.right.Set(!p.right.Get())
	p.dodge.Set(p.dodge.Get() + 1)
}

func (p *edgesPage) dodgeRow() string {
	if p.right.Get() {
		return "flex justify-end w-40"
	}
	return "flex justify-start w-40"
}

templ (p *edgesPage) Render() {
	<div class="flex-col">
		<span class="font-bold">Dodge</span>
		<div class={p.dodgeRow() + " my-1"}>
			<span class="px-1 border-rounded hover:border-red hover:text-red" onMouseEnter={p.moveAway} onMouseLeave={p.d.logged("leave dodge button")}>catch me</span>
		</div>
		<span class="font-dim">Entering the button moves it to the other side. It gets its</span>
		<span class="font-dim">leave right after the render, with the pointer still.</span>
		<span class="text-cyan mt-1">{fmt.Sprintf("dodges: %d", p.dodge.Get())}</span>
	</div>
}
