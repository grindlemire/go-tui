package main

import (
	"fmt"
	"time"
	tui "github.com/grindlemire/go-tui"
)

type watcherApp struct {
	// Stopwatch
	stopwatchSec *tui.State[int]
	stopwatchOn  *tui.State[bool]

	// Countdown
	countdownSec *tui.State[int]
	countdownOn  *tui.State[bool]

	// Live feed
	messages *tui.State[[]string]
	msgCh    chan string
	msgCount *tui.State[int]
}

func WatcherApp() *watcherApp {
	msgCh := make(chan string, 100)
	app := &watcherApp{
		stopwatchSec: tui.NewState(0),
		stopwatchOn:  tui.NewState(false),
		countdownSec: tui.NewState(300),
		countdownOn:  tui.NewState(false),
		messages:     tui.NewState([]string{}),
		msgCh:        msgCh,
		msgCount:     tui.NewState(0),
	}

	// Produce messages in background
	go func() {
		msgs := []string{
			"Hello from producer",
			"Data packet received",
			"Processing complete",
			"Checkpoint saved",
			"Heartbeat OK",
			"Cache refreshed",
			"Sync complete",
		}
		i := 0
		n := len(msgs)
		for {
			time.Sleep(3 * time.Second)
			msgCh <- msgs[i]
			i++
			if i >= n {
				i = 0
			}
		}
	}()

	return app
}

func (w *watcherApp) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnRune('q', func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnRune('s', func(ke tui.KeyEvent) { w.stopwatchOn.Set(!w.stopwatchOn.Get()) }),
		tui.OnRune('c', func(ke tui.KeyEvent) { w.countdownOn.Set(!w.countdownOn.Get()) }),
		tui.OnRune('r', func(ke tui.KeyEvent) {
			w.stopwatchSec.Set(0)
			w.countdownSec.Set(300)
			w.stopwatchOn.Set(false)
			w.countdownOn.Set(false)
		}),
	}
}

func (w *watcherApp) Watchers() []tui.Watcher {
	return []tui.Watcher{
		tui.OnTimer(time.Second, w.tick),
		tui.Watch(w.msgCh, w.addMessage),
	}
}

func (w *watcherApp) tick() {
	if w.stopwatchOn.Get() {
		w.stopwatchSec.Set(w.stopwatchSec.Get() + 1)
	}
	if w.countdownOn.Get() && w.countdownSec.Get() > 0 {
		w.countdownSec.Set(w.countdownSec.Get() - 1)
	}
}

func (w *watcherApp) addMessage(msg string) {
	current := w.messages.Get()
	ts := time.Now().Format("15:04:05")
	w.msgCount.Set(w.msgCount.Get() + 1)
	entry := fmt.Sprintf("[%s] Message #%d: %s", ts, w.msgCount.Get(), msg)
	// Keep last 10 messages
	if len(current) >= 10 {
		current = current[1:]
	}
	w.messages.Set(append(current, entry))
}

func formatDuration(seconds int) string {
	m := seconds / 60
	s := seconds - m*60
	return fmt.Sprintf("%02d:%02d", m, s)
}

templ (w *watcherApp) Render() {
	<div class="flex-col p-2 gap-2 border-rounded border-cyan">
		<span class="text-gradient-cyan-magenta font-bold">Timers & Watchers</span>

		// Top row: Stopwatch + Countdown
		<div class="flex gap-2">
			<div class="flex-col border-rounded p-1 gap-1 items-center" flexGrow={1.0}>
				<span class="text-gradient-cyan-magenta font-bold">Stopwatch</span>
				<br />
				<span class="text-cyan font-bold">{formatDuration(w.stopwatchSec.Get())}</span>
				<br />
				<div class="flex gap-1">
					@if w.stopwatchOn.Get() {
						<span class="text-green font-bold">Running</span>
					} @else {
						<span class="text-yellow">Paused</span>
					}
				</div>
				<span class="font-dim">[s] toggle</span>
			</div>

			<div class="flex-col border-rounded p-1 gap-1 items-center" flexGrow={1.0}>
				<span class="text-gradient-cyan-magenta font-bold">Countdown</span>
				<br />
				@if w.countdownSec.Get() == 0 {
					<span class="text-red font-bold">00:00</span>
				} @else {
					<span class="text-cyan font-bold">{formatDuration(w.countdownSec.Get())}</span>
				}
				<br />
				<div class="flex gap-1">
					@if w.countdownOn.Get() {
						<span class="text-green font-bold">Running</span>
					} @else {
						<span class="text-yellow">Paused</span>
					}
				</div>
				<span class="font-dim">[c] toggle</span>
			</div>
		</div>

		// Live Feed
		<div class="flex-col border-rounded p-1 gap-1">
			<span class="text-gradient-cyan-magenta font-bold">Live Feed</span>
			@for _, msg := range w.messages.Get() {
				<span class="text-green">{msg}</span>
			}
			@if len(w.messages.Get()) == 0 {
				<span class="font-dim">Waitingmessages...</span>
			}
		</div>

		<div class="flex gap-2 justify-center">
			<span class="font-dim">{fmt.Sprintf("Events: %d received", w.msgCount.Get())}</span>
		</div>

		<div class="flex justify-center">
			<span class="font-dim">s stopwatch | c countdown | r reset | q quit</span>
		</div>
	</div>
}
