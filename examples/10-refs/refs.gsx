package main

import (
	"fmt"
	tui "github.com/grindlemire/go-tui"
)

type refsApp struct {
	count        *tui.State[int]
	counter      *tui.Ref
	incrementBtn *tui.Ref
	decrementBtn *tui.Ref
	status       *tui.Ref
}

func Refs() *refsApp {
	return &refsApp{
		count:        tui.NewState(0),
		counter:      tui.NewRef(),
		incrementBtn: tui.NewRef(),
		decrementBtn: tui.NewRef(),
		status:       tui.NewRef(),
	}
}

func (r *refsApp) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnRune('q', func(ke tui.KeyEvent) { tui.Stop() }),
		tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { tui.Stop() }),
	}
}

func (r *refsApp) HandleMouse(me tui.MouseEvent) bool {
	if me.Button == tui.MouseLeft && me.Action == tui.MousePress {
		if r.incrementBtn.El() != nil && r.incrementBtn.El().ContainsPoint(me.X, me.Y) {
			r.count.Set(r.count.Get() + 1)
			return true
		}
		if r.decrementBtn.El() != nil && r.decrementBtn.El().ContainsPoint(me.X, me.Y) {
			r.count.Set(r.count.Get() - 1)
			return true
		}
	}
	return false
}

templ (r *refsApp) Render() {
	counter := r.counter
	incrementBtn := r.incrementBtn
	decrementBtn := r.decrementBtn
	status := r.status
	<div class="flex-col gap-1 p-2 border-rounded">
		<span class="font-bold text-cyan">Named Element References</span>
		<hr class="border" />
		<div ref={counter} class="border-single p-1">
			<span>
				Counter
				{fmt.Sprintf("%d", r.count.Get())}
			</span>
		</div>
		<div class="flex gap-1 w-full justify-center">
			<button ref={incrementBtn} class="border-single text-center p-1 w-10 h-5">{" + "}</button>
			<button ref={decrementBtn} class="border-single text-center p-1 w-10 h-5">{" - "}</button>
		</div>
		<div ref={status} class="font-dim">
			<span>Click buttons to update the counter</span>
		</div>
		<span class="font-dim">Press q to quit</span>
	</div>
}
