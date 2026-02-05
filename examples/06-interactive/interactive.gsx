package main

import (
	"fmt"
	"time"
	tui "github.com/grindlemire/go-tui"
)

type interactiveApp struct {
	count      *tui.State[int]
	elapsed    *tui.State[int]
	running    *tui.State[bool]
	lastEvent  *tui.State[string]
	eventCount *tui.State[int]
	sound      *tui.State[bool]
	notify     *tui.State[bool]
	dark       *tui.State[bool]
}

func Interactive() *interactiveApp {
	return &interactiveApp{
		count:      tui.NewState(0),
		elapsed:    tui.NewState(0),
		running:    tui.NewState(true),
		lastEvent:  tui.NewState("(waiting)"),
		eventCount: tui.NewState(0),
		sound:      tui.NewState(true),
		notify:     tui.NewState(false),
		dark:       tui.NewState(false),
	}
}

func (a *interactiveApp) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnRune('q', func(ke tui.KeyEvent) { tui.Stop() }),
		tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { tui.Stop() }),
		// Counter
		tui.OnRune('+', func(ke tui.KeyEvent) {
			a.count.Set(a.count.Get() + 1)
			a.eventCount.Set(a.eventCount.Get() + 1)
		}),
		tui.OnRune('=', func(ke tui.KeyEvent) {
			a.count.Set(a.count.Get() + 1)
			a.eventCount.Set(a.eventCount.Get() + 1)
		}),
		tui.OnRune('-', func(ke tui.KeyEvent) {
			a.count.Set(a.count.Get() - 1)
			a.eventCount.Set(a.eventCount.Get() + 1)
		}),
		tui.OnRune('0', func(ke tui.KeyEvent) {
			a.count.Set(0)
			a.eventCount.Set(a.eventCount.Get() + 1)
		}),
		// Timer
		tui.OnRune(' ', func(ke tui.KeyEvent) {
			a.running.Set(!a.running.Get())
		}),
		tui.OnRune('r', func(ke tui.KeyEvent) {
			a.elapsed.Set(0)
			a.eventCount.Set(a.eventCount.Get() + 1)
		}),
		// Toggles
		tui.OnRune('1', func(ke tui.KeyEvent) {
			a.sound.Set(!a.sound.Get())
			a.eventCount.Set(a.eventCount.Get() + 1)
		}),
		tui.OnRune('2', func(ke tui.KeyEvent) {
			a.notify.Set(!a.notify.Get())
			a.eventCount.Set(a.eventCount.Get() + 1)
		}),
		tui.OnRune('3', func(ke tui.KeyEvent) {
			a.dark.Set(!a.dark.Get())
			a.eventCount.Set(a.eventCount.Get() + 1)
		}),
	}
}

func (a *interactiveApp) timerTick() {
	if a.running.Get() {
		a.elapsed.Set(a.elapsed.Get() + 1)
	}
}

func (a *interactiveApp) increment(el *tui.Element) {
	a.count.Set(a.count.Get() + 1)
	a.eventCount.Set(a.eventCount.Get() + 1)
}

func (a *interactiveApp) decrement(el *tui.Element) {
	a.count.Set(a.count.Get() - 1)
	a.eventCount.Set(a.eventCount.Get() + 1)
}

func (a *interactiveApp) resetCount(el *tui.Element) {
	a.count.Set(0)
	a.eventCount.Set(a.eventCount.Get() + 1)
}

func (a *interactiveApp) toggleState(state *tui.State[bool]) func(*tui.Element) {
	return func(el *tui.Element) {
		state.Set(!state.Get())
		a.eventCount.Set(a.eventCount.Get() + 1)
	}
}

templ (a *interactiveApp) Render() {
	<div
		class="flex-col p-1 border-rounded gap-1"
		onEvent={a.inspectEvent}
		onTimer={tui.OnTimer(time.Second, a.timerTick)}>
		<div class="flex justify-between">
			<span class="text-gradient-cyan-magenta font-bold">{"Interactive Elements"}</span>
			<span class="text-blue font-bold">{fmt.Sprintf("Events: %d", a.eventCount.Get())}</span>
		</div>
		<div class="flex gap-1">
			<div
				class="border-single p-1 flex-col gap-1"
				flexGrow={1.0}>
				<span class="text-gradient-cyan-blue font-bold">{"onClick + KeyMap"}</span>
				<div class="flex gap-1 items-center">
					<span class="font-dim">Count:</span>
					<span class="text-cyan font-bold">{fmt.Sprintf("%d", a.count.Get())}</span>
				</div>
				<div class="flex gap-1">
					<button onClick={a.decrement}>{" - "}</button>
					<button onClick={a.increment}>{" + "}</button>
					<button onClick={a.resetCount}>{" 0 "}</button>
				</div>
				@if a.count.Get() > 0 {
					<span class="text-green font-bold">{"Positive"}</span>
				} @else @if a.count.Get() < 0 {
					<span class="text-red font-bold">{"Negative"}</span>
				} @else {
					<span class="text-blue font-bold">{"Zero"}</span>
				}
				<span class="font-dim">{"click btns or +/-/0"}</span>
			</div>
			<div
				class="border-single p-1 flex-col gap-1"
				flexGrow={1.0}>
				<span class="text-gradient-blue-cyan font-bold">{"onTimer"}</span>
				<div class="flex gap-1 items-center">
					<span class="font-dim">Elapsed:</span>
					<span class="text-blue font-bold">{formatTime(a.elapsed.Get())}</span>
				</div>
				@if a.running.Get() {
					<span class="text-green font-bold">{"Running"}</span>
				} @else {
					<span class="text-red font-bold">{"Stopped"}</span>
				}
				<span class="font-dim">{"[space] toggle [r] reset"}</span>
			</div>
		</div>
		<div class="flex gap-1">
			<div
				class="border-single p-1 flex-col gap-1"
				flexGrow={1.0}>
				<span class="text-gradient-green-cyan font-bold">{"onClick (toggles)"}</span>
				<div class="flex gap-1 items-center">
					<button onClick={a.toggleState(a.sound)}>{"Sound  "}</button>
					@if a.sound.Get() {
						<span class="text-green font-bold">ON</span>
					} @else {
						<span class="text-red font-bold">OFF</span>
					}
				</div>
				<div class="flex gap-1 items-center">
					<button onClick={a.toggleState(a.notify)}>{"Notify "}</button>
					@if a.notify.Get() {
						<span class="text-green font-bold">ON</span>
					} @else {
						<span class="text-red font-bold">OFF</span>
					}
				</div>
				<div class="flex gap-1 items-center">
					<button onClick={a.toggleState(a.dark)}>{"Theme  "}</button>
					@if a.dark.Get() {
						<span class="text-cyan font-bold">Dark</span>
					} @else {
						<span class="text-yellow font-bold">Light</span>
					}
				</div>
				<span class="font-dim">{"click or press 1/2/3"}</span>
			</div>
			<div
				class="border-single p-1 flex-col gap-1"
				flexGrow={1.0}>
				<span class="text-gradient-yellow-green font-bold">{"Event Inspector"}</span>
				<div class="flex gap-1 items-center">
					<span class="font-dim">Last:</span>
					<span class="text-yellow font-bold">{a.lastEvent.Get()}</span>
				</div>
				<div class="flex gap-1 items-center">
					<span class="font-dim">Total:</span>
					<span class="text-green font-bold">{fmt.Sprintf("%d", a.eventCount.Get())}</span>
				</div>
				<span class="font-dim">{"bubbled events shown"}</span>
			</div>
		</div>
		<div class="flex justify-between">
			<span class="font-dim">{"[q] quit"}</span>
		</div>
	</div>
}

func (a *interactiveApp) inspectEvent(el *tui.Element, e tui.Event) bool {
	a.eventCount.Set(a.eventCount.Get() + 1)
	switch ev := e.(type) {
	case tui.KeyEvent:
		if ev.Rune != 0 {
			a.lastEvent.Set(fmt.Sprintf("Key '%c'", ev.Rune))
		} else {
			a.lastEvent.Set(describeKey(ev.Key))
		}
	case tui.MouseEvent:
		a.lastEvent.Set(fmt.Sprintf("%s (%d,%d)", describeButton(ev.Button), ev.X, ev.Y))
	}
	return false
}

func formatTime(seconds int) string {
	m := seconds / 60
	s := seconds - (m * 60)
	return fmt.Sprintf("%02d:%02d", m, s)
}

func describeKey(key tui.Key) string {
	switch key {
	case tui.KeyEnter:
		return "Enter"
	case tui.KeyBackspace:
		return "Backspace"
	case tui.KeyTab:
		return "Tab"
	case tui.KeyUp:
		return "Up"
	case tui.KeyDown:
		return "Down"
	case tui.KeyLeft:
		return "Left"
	case tui.KeyRight:
		return "Right"
	case tui.KeyHome:
		return "Home"
	case tui.KeyEnd:
		return "End"
	default:
		return fmt.Sprintf("Key(%d)", key)
	}
}

func describeButton(btn tui.MouseButton) string {
	switch btn {
	case tui.MouseLeft:
		return "Left Click"
	case tui.MouseRight:
		return "Right Click"
	case tui.MouseMiddle:
		return "Middle Click"
	case tui.MouseWheelUp:
		return "Wheel Up"
	case tui.MouseWheelDown:
		return "Wheel Down"
	default:
		return "Mouse"
	}
}
