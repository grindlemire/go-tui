package main

import (
	"fmt"
	tui "github.com/grindlemire/go-tui"
)

type explorer struct {
	lastKey  *tui.State[string]
	keyCount *tui.State[int]
}

func Explorer() *explorer {
	return &explorer{
		lastKey:  tui.NewState("(none)"),
		keyCount: tui.NewState(0),
	}
}

func (e *explorer) record(name string) {
	e.keyCount.Set(e.keyCount.Get() + 1)
	e.lastKey.Set(name)
}

func (e *explorer) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnStop(tui.Rune('q'), func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.On(tui.AnyRune, func(ke tui.KeyEvent) {
			e.record(fmt.Sprintf("'%c' (rune)", ke.Rune))
		}),
		// With Kitty keyboard protocol, Ctrl+H/I/M arrive as KeyRune
		// events with ModCtrl and are matched by On with Rune('x').Ctrl().
		// Without Kitty, they are indistinguishable from Backspace/Tab/Enter
		// and match the On handlers below instead.
		tui.On(tui.Rune('h').Ctrl(), func(ke tui.KeyEvent) { e.record("Ctrl+'h' (rune)") }),
		tui.On(tui.Rune('i').Ctrl(), func(ke tui.KeyEvent) { e.record("Ctrl+'i' (rune)") }),
		tui.On(tui.Rune('m').Ctrl(), func(ke tui.KeyEvent) { e.record("Ctrl+'m' (rune)") }),
		tui.On(tui.KeyEnter, func(ke tui.KeyEvent) { e.record("Enter") }),
		tui.On(tui.KeyTab, func(ke tui.KeyEvent) { e.record("Tab") }),
		tui.On(tui.KeyBackspace, func(ke tui.KeyEvent) { e.record("Backspace") }),
		tui.On(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.On(tui.KeyUp, func(ke tui.KeyEvent) { e.record("Up") }),
		tui.On(tui.KeyDown, func(ke tui.KeyEvent) { e.record("Down") }),
		tui.On(tui.KeyLeft, func(ke tui.KeyEvent) { e.record("Left") }),
		tui.On(tui.KeyRight, func(ke tui.KeyEvent) { e.record("Right") }),
		tui.On(tui.Rune('a').Ctrl(), func(ke tui.KeyEvent) { e.record("Ctrl+A") }),
		tui.On(tui.Rune('s').Ctrl(), func(ke tui.KeyEvent) { e.record("Ctrl+S") }),
	}
}

templ (e *explorer) Render() {
	<div class="flex-col gap-1 p-2 border-rounded border-cyan">
		<span class="text-gradient-cyan-magenta font-bold">Keyboard Explorer</span>
		<hr class="border-single" />
		<div class="flex gap-2">
			<span class="font-dim">Last Key:</span>
			<span class="text-cyan font-bold">{e.lastKey.Get()}</span>
		</div>
		<div class="flex gap-2">
			<span class="font-dim">Key Count:</span>
			<span class="text-cyan font-bold">{fmt.Sprintf("%d", e.keyCount.Get())}</span>
		</div>

		<br />
		<span class="font-dim">Press any key to see it displayed above</span>
		<span class="font-dim">Press q or Esc to quit</span>
	</div>
}
