package main

import (
	"fmt"
	tui "github.com/grindlemire/go-tui"
)

type keyboardApp struct {
	lastKey  *tui.State[string]
	keyCount *tui.State[int]
}

func Keyboard() *keyboardApp {
	return &keyboardApp{
		lastKey:  tui.NewState("(none)"),
		keyCount: tui.NewState(0),
	}
}

func (k *keyboardApp) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnRune('q', func(ke tui.KeyEvent) { tui.Stop() }),
		tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { tui.Stop() }),
		tui.OnRunes(func(ke tui.KeyEvent) {
			k.keyCount.Set(k.keyCount.Get() + 1)
			k.lastKey.Set(fmt.Sprintf("'%c' (rune)", ke.Rune))
		}),
		tui.OnKey(tui.KeyEnter, func(ke tui.KeyEvent) { k.recordSpecial("Enter") }),
		tui.OnKey(tui.KeyBackspace, func(ke tui.KeyEvent) { k.recordSpecial("Backspace") }),
		tui.OnKey(tui.KeyTab, func(ke tui.KeyEvent) { k.recordSpecial("Tab") }),
		tui.OnKey(tui.KeyUp, func(ke tui.KeyEvent) { k.recordSpecial("Up Arrow") }),
		tui.OnKey(tui.KeyDown, func(ke tui.KeyEvent) { k.recordSpecial("Down Arrow") }),
		tui.OnKey(tui.KeyLeft, func(ke tui.KeyEvent) { k.recordSpecial("Left Arrow") }),
		tui.OnKey(tui.KeyRight, func(ke tui.KeyEvent) { k.recordSpecial("Right Arrow") }),
		tui.OnKey(tui.KeyHome, func(ke tui.KeyEvent) { k.recordSpecial("Home") }),
		tui.OnKey(tui.KeyEnd, func(ke tui.KeyEvent) { k.recordSpecial("End") }),
		tui.OnKey(tui.KeyPageUp, func(ke tui.KeyEvent) { k.recordSpecial("Page Up") }),
		tui.OnKey(tui.KeyPageDown, func(ke tui.KeyEvent) { k.recordSpecial("Page Down") }),
		tui.OnKey(tui.KeyDelete, func(ke tui.KeyEvent) { k.recordSpecial("Delete") }),
		tui.OnKey(tui.KeyInsert, func(ke tui.KeyEvent) { k.recordSpecial("Insert") }),
	}
}

func (k *keyboardApp) recordSpecial(name string) {
	k.keyCount.Set(k.keyCount.Get() + 1)
	k.lastKey.Set(name)
}

templ (k *keyboardApp) Render() {
	<div class="flex-col gap-1 p-2 border-rounded">
		<span class="font-bold text-cyan">Keyboard Events</span>
		<hr class="border" />

		<div class="flex gap-2">
			<span>Last key pressed:</span>
			<span class="font-bold text-green">{k.lastKey.Get()}</span>
		</div>

		<div class="flex gap-2">
			<span>Total keys pressed:</span>
			<span class="font-bold text-blue">{fmt.Sprintf("%d", k.keyCount.Get())}</span>
		</div>

		<br />
		<span class="font-dim">Press any key to see it displayed, q to quit</span>
	</div>
}
