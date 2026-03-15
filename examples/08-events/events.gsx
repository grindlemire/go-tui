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
		tui.OnRuneStop('q', func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnRunes(func(ke tui.KeyEvent) {
			e.record(fmt.Sprintf("'%c' (rune)", ke.Rune))
		}),
		// With Kitty keyboard protocol, Ctrl+H/I/M arrive as KeyRune
		// events with ModCtrl and are matched by OnRuneMod.
		// Without Kitty, they are indistinguishable from Backspace/Tab/Enter
		// and match the OnKey handlers below instead.
		tui.OnRuneMod('h', tui.ModCtrl, func(ke tui.KeyEvent) { e.record("Ctrl+'h' (rune)") }),
		tui.OnRuneMod('i', tui.ModCtrl, func(ke tui.KeyEvent) { e.record("Ctrl+'i' (rune)") }),
		tui.OnRuneMod('m', tui.ModCtrl, func(ke tui.KeyEvent) { e.record("Ctrl+'m' (rune)") }),
		tui.OnKey(tui.KeyEnter, func(ke tui.KeyEvent) { e.record("Enter") }),
		tui.OnKey(tui.KeyTab, func(ke tui.KeyEvent) { e.record("Tab") }),
		tui.OnKey(tui.KeyBackspace, func(ke tui.KeyEvent) { e.record("Backspace") }),
		tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnKey(tui.KeyUp, func(ke tui.KeyEvent) { e.record("Up") }),
		tui.OnKey(tui.KeyDown, func(ke tui.KeyEvent) { e.record("Down") }),
		tui.OnKey(tui.KeyLeft, func(ke tui.KeyEvent) { e.record("Left") }),
		tui.OnKey(tui.KeyRight, func(ke tui.KeyEvent) { e.record("Right") }),
		tui.OnKey(tui.KeyCtrlA, func(ke tui.KeyEvent) { e.record("Ctrl+A") }),
		tui.OnKey(tui.KeyCtrlS, func(ke tui.KeyEvent) { e.record("Ctrl+S") }),
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
