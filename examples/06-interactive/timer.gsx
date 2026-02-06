package main

import (
	"fmt"
	"time"
	tui "github.com/grindlemire/go-tui"
)

type timer struct {
	elapsed *tui.State[int]
	running *tui.State[bool]
	events  *Events[string]
}

func Timer(events *Events[string]) *timer {
	return &timer{
		elapsed: tui.NewState(0),
		running: tui.NewState(true),
		events:  events,
	}
}

func (t *timer) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnRune(' ', func(ke tui.KeyEvent) { t.toggleRunning() }),
		tui.OnRune('r', func(ke tui.KeyEvent) { t.resetTimer() }),
	}
}

func (t *timer) Watchers() []tui.Watcher {
	return []tui.Watcher{
		tui.OnTimer(time.Second, t.tick),
	}
}

func (t *timer) toggleRunning() {
	t.running.Set(!t.running.Get())
	t.events.Emit("timer toggle")
}

func (t *timer) resetTimer() {
	t.elapsed.Set(0)
	t.events.Emit("timer reset")
}

func (t *timer) tick() {
	if t.running.Get() {
		t.elapsed.Set(t.elapsed.Get() + 1)
	}
}

func formatTime(seconds int) string {
	m := seconds / 60
	s := seconds - (m * 60)
	return fmt.Sprintf("%02d:%02d", m, s)
}

templ (t *timer) Render() {
	<div class="border-single p-1 flex-col gap-1" flexGrow={1.0}>
		<span class="text-gradient-blue-cyan font-bold">{"Timer"}</span>
		<div class="flex gap-1 items-center">
			<span class="font-dim">Elapsed:</span>
			<span class="text-blue font-bold">{formatTime(t.elapsed.Get())}</span>
		</div>
		@if t.running.Get() {
			<span class="text-green font-bold">{"Running"}</span>
		} @else {
			<span class="text-red font-bold">{"Stopped"}</span>
		}
		<span class="font-dim">{"[space] toggle [r] reset"}</span>
	</div>
}
