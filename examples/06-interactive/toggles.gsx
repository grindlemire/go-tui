package main

import tui "github.com/grindlemire/go-tui"

type toggles struct {
	sound     *tui.State[bool]
	notify    *tui.State[bool]
	dark      *tui.State[bool]
	events    *Events[string]
	soundBtn  *tui.Ref
	notifyBtn *tui.Ref
	themeBtn  *tui.Ref
}

func Toggles(events *Events[string]) *toggles {
	return &toggles{
		sound:     tui.NewState(true),
		notify:    tui.NewState(false),
		dark:      tui.NewState(false),
		events:    events,
		soundBtn:  tui.NewRef(),
		notifyBtn: tui.NewRef(),
		themeBtn:  tui.NewRef(),
	}
}

func (t *toggles) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnRune('1', func(ke tui.KeyEvent) { t.toggleSound() }),
		tui.OnRune('2', func(ke tui.KeyEvent) { t.toggleNotify() }),
		tui.OnRune('3', func(ke tui.KeyEvent) { t.toggleTheme() }),
	}
}

func (t *toggles) HandleMouse(me tui.MouseEvent) bool {
	return tui.HandleClicks(me,
		tui.Click(t.soundBtn, t.toggleSound),
		tui.Click(t.notifyBtn, t.toggleNotify),
		tui.Click(t.themeBtn, t.toggleTheme),
	)
}

func (t *toggles) toggleSound() {
	t.sound.Set(!t.sound.Get())
	t.events.Emit("toggle sound")
}

func (t *toggles) toggleNotify() {
	t.notify.Set(!t.notify.Get())
	t.events.Emit("toggle notify")
}

func (t *toggles) toggleTheme() {
	t.dark.Set(!t.dark.Get())
	t.events.Emit("toggle theme")
}

templ (t *toggles) Render() {
	<div class="border-single p-1 flex-col gap-1" flexGrow={1.0}>
		<span class="text-gradient-green-cyan font-bold">{"Toggles"}</span>
		<div class="flex gap-1 items-center">
			<button ref={t.soundBtn}>{"Sound  "}</button>
			@if t.sound.Get() {
				<span class="text-green font-bold">ON</span>
			} @else {
				<span class="text-red font-bold">OFF</span>
			}
		</div>
		<div class="flex gap-1 items-center">
			<button ref={t.notifyBtn}>{"Notify "}</button>
			@if t.notify.Get() {
				<span class="text-green font-bold">ON</span>
			} @else {
				<span class="text-red font-bold">OFF</span>
			}
		</div>
		<div class="flex gap-1 items-center">
			<button ref={t.themeBtn}>{"Theme  "}</button>
			@if t.dark.Get() {
				<span class="text-cyan font-bold">Dark</span>
			} @else {
				<span class="text-yellow font-bold">Light</span>
			}
		</div>
		<span class="font-dim">{"click or press 1/2/3"}</span>
	</div>
}
