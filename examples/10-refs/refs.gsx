package main

import (
	"fmt"
	tui "github.com/grindlemire/go-tui"
)

type refsApp struct {
	count        *tui.State[int]
	incrementBtn *tui.Ref
	decrementBtn *tui.Ref
	resetBtn     *tui.Ref
}

func Refs() *refsApp {
	return &refsApp{
		count:        tui.NewState(0),
		incrementBtn: tui.NewRef(),
		decrementBtn: tui.NewRef(),
		resetBtn:     tui.NewRef(),
	}
}

func (r *refsApp) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnRune('q', func(ke tui.KeyEvent) { tui.Stop() }),
		tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { tui.Stop() }),
		tui.OnRune('+', func(ke tui.KeyEvent) { r.increment() }),
		tui.OnRune('=', func(ke tui.KeyEvent) { r.increment() }),
		tui.OnRune('-', func(ke tui.KeyEvent) { r.decrement() }),
		tui.OnRune('0', func(ke tui.KeyEvent) { r.reset() }),
	}
}

func (r *refsApp) HandleMouse(me tui.MouseEvent) bool {
	return tui.HandleClicks(me,
		tui.Click(r.incrementBtn, r.increment),
		tui.Click(r.decrementBtn, r.decrement),
		tui.Click(r.resetBtn, r.reset),
	)
}

func (r *refsApp) increment() {
	r.count.Set(r.count.Get() + 1)
}

func (r *refsApp) decrement() {
	r.count.Set(r.count.Get() - 1)
}

func (r *refsApp) reset() {
	r.count.Set(0)
}

func (r *refsApp) countStyle() tui.Style {
	c := r.count.Get()
	if c > 0 {
		return tui.NewStyle().Bold().Foreground(tui.Green)
	} else if c < 0 {
		return tui.NewStyle().Bold().Foreground(tui.Red)
	}
	return tui.NewStyle().Bold().Foreground(tui.Blue)
}

templ (r *refsApp) Render() {
	<div class="flex-col gap-1 p-2 border-rounded justify-center items-center h-full">
		<span class="text-gradient-cyan-magenta font-bold">{"Element References Demo"}</span>
		<hr class="border w-full" />
		<div class="border-single p-2 flex-col items-center gap-1">
			<span class="font-dim">{"Count"}</span>
			<span textStyle={r.countStyle()}>{fmt.Sprintf("%d", r.count.Get())}</span>
		</div>
		<div class="flex gap-2 justify-center">
			<button ref={r.decrementBtn}>{" - "}</button>
			<button ref={r.resetBtn}>{" 0 "}</button>
			<button ref={r.incrementBtn}>{" + "}</button>
		</div>
		@if r.count.Get() > 0 {
			<span class="text-green font-bold">{"Positive"}</span>
		} @else @if r.count.Get() < 0 {
			<span class="text-red font-bold">{"Negative"}</span>
		} @else {
			<span class="text-blue font-bold">{"Zero"}</span>
		}
		<span class="font-dim">{"[+/-/0] keys or click buttons | [q] quit"}</span>
	</div>
}
