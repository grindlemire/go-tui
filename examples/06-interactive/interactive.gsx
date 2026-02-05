package main

import (
	"fmt"
	"time"
	tui "github.com/grindlemire/go-tui"
)

templ Interactive() {
	count := tui.NewState(0)
	elapsed := tui.NewState(0)
	running := tui.NewState(true)
	lastEvent := tui.NewState("(waiting)")
	eventCount := tui.NewState(0)
	sound := tui.NewState(true)
	notify := tui.NewState(false)
	dark := tui.NewState(false)
	<div
		class="flex-col p-1 border-rounded gap-1"
		onKeyPress={handleKeys(count, running, elapsed, sound, notify, dark, eventCount)}
		onEvent={inspectEvent(lastEvent, eventCount)}
		onTimer={tui.OnTimer(time.Second, timerTick(elapsed, running))}>
		<div class="flex justify-between">
			<span class="text-gradient-cyan-magenta font-bold">{"Interactive Elements"}</span>
			<span class="text-blue font-bold">{fmt.Sprintf("Events: %d", eventCount.Get())}</span>
		</div>
		<div class="flex gap-1">
			<div
				class="border-single p-1 flex-col gap-1"
				flexGrow={1.0}>
				<span class="text-gradient-cyan-blue font-bold">{"onClick + onKeyPress"}</span>
				<div class="flex gap-1 items-center">
					<span class="font-dim">Count:</span>
					<span class="text-cyan font-bold">{fmt.Sprintf("%d", count.Get())}</span>
				</div>
				<div class="flex gap-1">
					<button onClick={decrement(count, eventCount)}>{" - "}</button>
					<button onClick={increment(count, eventCount)}>{" + "}</button>
					<button onClick={resetCount(count, eventCount)}>{" 0 "}</button>
				</div>
				@if count.Get() > 0 {
					<span class="text-green font-bold">{"Positive"}</span>
				} @else @if count.Get() < 0 {
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
					<span class="text-blue font-bold">{formatTime(elapsed.Get())}</span>
				</div>
				@if running.Get() {
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
					<button onClick={toggle(sound, eventCount)}>{"Sound  "}</button>
					@if sound.Get() {
						<span class="text-green font-bold">ON</span>
					} @else {
						<span class="text-red font-bold">OFF</span>
					}
				</div>
				<div class="flex gap-1 items-center">
					<button onClick={toggle(notify, eventCount)}>{"Notify "}</button>
					@if notify.Get() {
						<span class="text-green font-bold">ON</span>
					} @else {
						<span class="text-red font-bold">OFF</span>
					}
				</div>
				<div class="flex gap-1 items-center">
					<button onClick={toggle(dark, eventCount)}>{"Theme  "}</button>
					@if dark.Get() {
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
					<span class="text-yellow font-bold">{lastEvent.Get()}</span>
				</div>
				<div class="flex gap-1 items-center">
					<span class="font-dim">Total:</span>
					<span class="text-green font-bold">{fmt.Sprintf("%d", eventCount.Get())}</span>
				</div>
				<span class="font-dim">{"bubbled events shown"}</span>
			</div>
		</div>
		<div class="flex justify-between">
			<span class="font-dim">{"[q] quit"}</span>
		</div>
	</div>
}

func increment(count *tui.State[int], eventCount *tui.State[int]) func(*tui.Element) {
	return func(el *tui.Element) {
		count.Set(count.Get() + 1)
		eventCount.Set(eventCount.Get() + 1)
	}
}

func decrement(count *tui.State[int], eventCount *tui.State[int]) func(*tui.Element) {
	return func(el *tui.Element) {
		count.Set(count.Get() - 1)
		eventCount.Set(eventCount.Get() + 1)
	}
}

func resetCount(count *tui.State[int], eventCount *tui.State[int]) func(*tui.Element) {
	return func(el *tui.Element) {
		count.Set(0)
		eventCount.Set(eventCount.Get() + 1)
	}
}

func handleKeys(count *tui.State[int], running *tui.State[bool], elapsed *tui.State[int], sound, notify, dark *tui.State[bool], eventCount *tui.State[int]) func(*tui.Element, tui.KeyEvent) bool {
	return func(el *tui.Element, e tui.KeyEvent) bool {
		switch e.Rune {
		// Counter
		case '+', '=':
			count.Set(count.Get() + 1)
			eventCount.Set(eventCount.Get() + 1)
			return true
		case '-':
			count.Set(count.Get() - 1)
			eventCount.Set(eventCount.Get() + 1)
			return true
		case '0':
			count.Set(0)
			eventCount.Set(eventCount.Get() + 1)
			return true
		// Timer
		case ' ':
			running.Set(!running.Get())
			return true
		case 'r':
			elapsed.Set(0)
			eventCount.Set(eventCount.Get() + 1)
			return true
		// Toggles
		case '1':
			sound.Set(!sound.Get())
			eventCount.Set(eventCount.Get() + 1)
			return true
		case '2':
			notify.Set(!notify.Get())
			eventCount.Set(eventCount.Get() + 1)
			return true
		case '3':
			dark.Set(!dark.Get())
			eventCount.Set(eventCount.Get() + 1)
			return true
		}
		return false
	}
}

func timerTick(elapsed *tui.State[int], running *tui.State[bool]) func() {
	return func() {
		if running.Get() {
			elapsed.Set(elapsed.Get() + 1)
		}
	}
}

func formatTime(seconds int) string {
	m := seconds / 60
	s := seconds - (m * 60)
	return fmt.Sprintf("%02d:%02d", m, s)
}

func toggle(state *tui.State[bool], eventCount *tui.State[int]) func(*tui.Element) {
	return func(el *tui.Element) {
		state.Set(!state.Get())
		eventCount.Set(eventCount.Get() + 1)
	}
}

func inspectEvent(lastEvent *tui.State[string], eventCount *tui.State[int]) func(*tui.Element, tui.Event) bool {
	return func(el *tui.Element, e tui.Event) bool {
		eventCount.Set(eventCount.Get() + 1)
		switch ev := e.(type) {
		case tui.KeyEvent:
			if ev.Rune != 0 {
				lastEvent.Set(fmt.Sprintf("Key '%c'", ev.Rune))
			} else {
				lastEvent.Set(describeKey(ev.Key))
			}
		case tui.MouseEvent:
			lastEvent.Set(fmt.Sprintf("%s (%d,%d)", describeButton(ev.Button), ev.X, ev.Y))
		}
		return false
	}
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
