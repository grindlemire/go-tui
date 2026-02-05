package main

import (
	"fmt"
	tui "github.com/grindlemire/go-tui"
)

templ Keyboard() {
	lastKey := tui.NewState("(none)")
	keyCount := tui.NewState(0)
	<div class="flex-col gap-1 p-2 border-rounded"
	     onKeyPress={handleKey(lastKey, keyCount)}
	     focusable={true}>
		<span class="font-bold text-cyan">Keyboard Events</span>
		<hr class="border" />

		<div class="flex gap-2">
			<span>Last key pressed:</span>
			<span class="font-bold text-green">{lastKey.Get()}</span>
		</div>

		<div class="flex gap-2">
			<span>Total keys pressed:</span>
			<span class="font-bold text-blue">{fmt.Sprintf("%d", keyCount.Get())}</span>
		</div>

		<br />
		<span class="font-dim">Press any key to see it displayed, q to quit</span>
	</div>
}

func handleKey(lastKey *tui.State[string], keyCount *tui.State[int]) func(*tui.Element, tui.KeyEvent) bool {
	return func(el *tui.Element, e tui.KeyEvent) bool {
		keyCount.Set(keyCount.Get() + 1)

		if e.Rune != 0 {
			lastKey.Set(fmt.Sprintf("'%c' (rune)", e.Rune))
		} else {
			lastKey.Set(keyName(e.Key))
		}
		return true
	}
}

func keyName(key tui.Key) string {
	switch key {
	case tui.KeyEnter:
		return "Enter"
	case tui.KeyBackspace:
		return "Backspace"
	case tui.KeyTab:
		return "Tab"
	case tui.KeyEscape:
		return "Escape"
	case tui.KeyUp:
		return "Up Arrow"
	case tui.KeyDown:
		return "Down Arrow"
	case tui.KeyLeft:
		return "Left Arrow"
	case tui.KeyRight:
		return "Right Arrow"
	case tui.KeyHome:
		return "Home"
	case tui.KeyEnd:
		return "End"
	case tui.KeyPageUp:
		return "Page Up"
	case tui.KeyPageDown:
		return "Page Down"
	case tui.KeyDelete:
		return "Delete"
	case tui.KeyInsert:
		return "Insert"
	default:
		return fmt.Sprintf("Key(%d)", key)
	}
}
