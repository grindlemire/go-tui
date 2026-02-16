package main

import (
	"fmt"
	tui "github.com/grindlemire/go-tui"
)

type keyboardApp struct {
	lastKey      *tui.State[string]
	keyCount     *tui.State[int]
	pressedEnter *tui.State[bool]
	pressedTab   *tui.State[bool]
	pressedBksp  *tui.State[bool]
	pressedDel   *tui.State[bool]
	pressedUp    *tui.State[bool]
	pressedDown  *tui.State[bool]
	pressedLeft  *tui.State[bool]
	pressedRight *tui.State[bool]
	pressedHome  *tui.State[bool]
	pressedEnd   *tui.State[bool]
	pressedPgUp  *tui.State[bool]
	pressedPgDn  *tui.State[bool]
	pressedCtrlA *tui.State[bool]
	pressedCtrlD *tui.State[bool]
	pressedCtrlN *tui.State[bool]
	pressedCtrlR *tui.State[bool]
	pressedCtrlS *tui.State[bool]
	pressedCtrlW *tui.State[bool]
	pressedCtrlX *tui.State[bool]
	pressedCtrlZ *tui.State[bool]
}

func Keyboard() *keyboardApp {
	return &keyboardApp{
		lastKey:      tui.NewState("(none)"),
		keyCount:     tui.NewState(0),
		pressedEnter: tui.NewState(false),
		pressedTab:   tui.NewState(false),
		pressedBksp:  tui.NewState(false),
		pressedDel:   tui.NewState(false),
		pressedUp:    tui.NewState(false),
		pressedDown:  tui.NewState(false),
		pressedLeft:  tui.NewState(false),
		pressedRight: tui.NewState(false),
		pressedHome:  tui.NewState(false),
		pressedEnd:   tui.NewState(false),
		pressedPgUp:  tui.NewState(false),
		pressedPgDn:  tui.NewState(false),
		pressedCtrlA: tui.NewState(false),
		pressedCtrlD: tui.NewState(false),
		pressedCtrlN: tui.NewState(false),
		pressedCtrlR: tui.NewState(false),
		pressedCtrlS: tui.NewState(false),
		pressedCtrlW: tui.NewState(false),
		pressedCtrlX: tui.NewState(false),
		pressedCtrlZ: tui.NewState(false),
	}
}

func (k *keyboardApp) record(name string) {
	k.keyCount.Set(k.keyCount.Get() + 1)
	k.lastKey.Set(name)
}

func (k *keyboardApp) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnRuneStop('q', func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnRunes(func(ke tui.KeyEvent) {
			k.record(fmt.Sprintf("'%c' (rune)", ke.Rune))
		}),
		tui.OnKey(tui.KeyEnter, func(ke tui.KeyEvent) { k.pressedEnter.Set(true); k.record("Enter") }),
		tui.OnKey(tui.KeyTab, func(ke tui.KeyEvent) { k.pressedTab.Set(true); k.record("Tab") }),
		tui.OnKey(tui.KeyBackspace, func(ke tui.KeyEvent) { k.pressedBksp.Set(true); k.record("Backspace") }),
		tui.OnKey(tui.KeyDelete, func(ke tui.KeyEvent) { k.pressedDel.Set(true); k.record("Delete") }),
		tui.OnKey(tui.KeyUp, func(ke tui.KeyEvent) { k.pressedUp.Set(true); k.record("Up") }),
		tui.OnKey(tui.KeyDown, func(ke tui.KeyEvent) { k.pressedDown.Set(true); k.record("Down") }),
		tui.OnKey(tui.KeyLeft, func(ke tui.KeyEvent) { k.pressedLeft.Set(true); k.record("Left") }),
		tui.OnKey(tui.KeyRight, func(ke tui.KeyEvent) { k.pressedRight.Set(true); k.record("Right") }),
		tui.OnKey(tui.KeyHome, func(ke tui.KeyEvent) { k.pressedHome.Set(true); k.record("Home") }),
		tui.OnKey(tui.KeyEnd, func(ke tui.KeyEvent) { k.pressedEnd.Set(true); k.record("End") }),
		tui.OnKey(tui.KeyPageUp, func(ke tui.KeyEvent) { k.pressedPgUp.Set(true); k.record("PgUp") }),
		tui.OnKey(tui.KeyPageDown, func(ke tui.KeyEvent) { k.pressedPgDn.Set(true); k.record("PgDn") }),
		tui.OnKey(tui.KeyCtrlA, func(ke tui.KeyEvent) { k.pressedCtrlA.Set(true); k.record("Ctrl+A") }),
		tui.OnKey(tui.KeyCtrlD, func(ke tui.KeyEvent) { k.pressedCtrlD.Set(true); k.record("Ctrl+D") }),
		tui.OnKey(tui.KeyCtrlN, func(ke tui.KeyEvent) { k.pressedCtrlN.Set(true); k.record("Ctrl+N") }),
		tui.OnKey(tui.KeyCtrlR, func(ke tui.KeyEvent) { k.pressedCtrlR.Set(true); k.record("Ctrl+R") }),
		tui.OnKey(tui.KeyCtrlS, func(ke tui.KeyEvent) { k.pressedCtrlS.Set(true); k.record("Ctrl+S") }),
		tui.OnKey(tui.KeyCtrlW, func(ke tui.KeyEvent) { k.pressedCtrlW.Set(true); k.record("Ctrl+W") }),
		tui.OnKey(tui.KeyCtrlX, func(ke tui.KeyEvent) { k.pressedCtrlX.Set(true); k.record("Ctrl+X") }),
		tui.OnKey(tui.KeyCtrlZ, func(ke tui.KeyEvent) { k.pressedCtrlZ.Set(true); k.record("Ctrl+Z") }),
	}
}

func keyLabel(label string, pressed bool) string {
	if pressed {
		return "✓ " + label
	}
	return "· " + label
}

func keyStyle(pressed bool) string {
	if pressed {
		return "text-green font-bold"
	}
	return "font-dim"
}

templ (k *keyboardApp) Render() {
	<div class="flex-col gap-1 p-2 border-rounded border-cyan">
		<span class="text-gradient-cyan-magenta font-bold">Keyboard Explorer</span>
		<hr class="border-single" />
		<div class="flex gap-2">
			<span class="font-dim">Last Key:</span>
			<span class="text-gradient-cyan-magenta font-bold">{k.lastKey.Get()}</span>
		</div>

		<div class="flex gap-2">
			<span class="font-dim">Key Count:</span>
			<span class="text-cyan font-bold">{fmt.Sprintf("%d", k.keyCount.Get())}</span>
		</div>

		<div class="flex-row justify-between">
			<div class="flex-col border-rounded p-1 gap-1">
				<span class="text-gradient-cyan-magenta font-bold">Special Keys</span>
				<div class="flex gap-2">
					<span class={keyStyle(k.pressedEnter.Get())}>{keyLabel("Enter", k.pressedEnter.Get())}</span>
					<span class={keyStyle(k.pressedTab.Get())}>{keyLabel("Tab", k.pressedTab.Get())}</span>
					<span class={keyStyle(k.pressedBksp.Get())}>{keyLabel("Bksp", k.pressedBksp.Get())}</span>
					<span class={keyStyle(k.pressedDel.Get())}>{keyLabel("Del", k.pressedDel.Get())}</span>
				</div>
				<div class="flex gap-2">
					<span class={keyStyle(k.pressedUp.Get())}>{keyLabel("Up", k.pressedUp.Get())}</span>
					<span class={keyStyle(k.pressedDown.Get())}>{keyLabel("Down", k.pressedDown.Get())}</span>
					<span class={keyStyle(k.pressedLeft.Get())}>{keyLabel("Left", k.pressedLeft.Get())}</span>
					<span class={keyStyle(k.pressedRight.Get())}>{keyLabel("Right", k.pressedRight.Get())}</span>
				</div>
				<div class="flex gap-2">
					<span class={keyStyle(k.pressedHome.Get())}>{keyLabel("Home", k.pressedHome.Get())}</span>
					<span class={keyStyle(k.pressedEnd.Get())}>{keyLabel("End", k.pressedEnd.Get())}</span>
					<span class={keyStyle(k.pressedPgUp.Get())}>{keyLabel("PgUp", k.pressedPgUp.Get())}</span>
					<span class={keyStyle(k.pressedPgDn.Get())}>{keyLabel("PgDn", k.pressedPgDn.Get())}</span>
				</div>
			</div>

			<div class="flex-col border-rounded p-1 gap-1">
				<span class="text-gradient-cyan-magenta font-bold">Ctrl Keys</span>
				<div class="flex gap-2">
					<span class={keyStyle(k.pressedCtrlA.Get())}>{keyLabel("Ctrl+A", k.pressedCtrlA.Get())}</span>
					<span class={keyStyle(k.pressedCtrlD.Get())}>{keyLabel("Ctrl+D", k.pressedCtrlD.Get())}</span>
					<span class={keyStyle(k.pressedCtrlN.Get())}>{keyLabel("Ctrl+N", k.pressedCtrlN.Get())}</span>
					<span class={keyStyle(k.pressedCtrlR.Get())}>{keyLabel("Ctrl+R", k.pressedCtrlR.Get())}</span>
				</div>
				<div class="flex gap-2">
					<span class={keyStyle(k.pressedCtrlS.Get())}>{keyLabel("Ctrl+S", k.pressedCtrlS.Get())}</span>
					<span class={keyStyle(k.pressedCtrlW.Get())}>{keyLabel("Ctrl+W", k.pressedCtrlW.Get())}</span>
					<span class={keyStyle(k.pressedCtrlX.Get())}>{keyLabel("Ctrl+X", k.pressedCtrlX.Get())}</span>
					<span class={keyStyle(k.pressedCtrlZ.Get())}>{keyLabel("Ctrl+Z", k.pressedCtrlZ.Get())}</span>
				</div>
			</div>
		</div>

		<br />
		<span class="font-dim">Press any key to see it displayed above</span>
		<span class="font-dim">Press q to quit</span>
	</div>
}
