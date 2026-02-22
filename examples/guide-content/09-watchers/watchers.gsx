package main

import (
	"fmt"
	"time"

	tui "github.com/grindlemire/go-tui"
)

type watcherApp struct {
	stopwatchSec *tui.State[int]
	stopwatchOn  *tui.State[bool]
	messages     *tui.State[[]string]
	msgCh        chan string
	msgCount     *tui.State[int]
	feed         *tui.Ref
	scrollY      *tui.State[int]
}

func WatcherApp() *watcherApp {
	msgCh := make(chan string, 100)
	app := &watcherApp{
		stopwatchSec: tui.NewState(0),
		stopwatchOn:  tui.NewState(false),
		messages:     tui.NewState([]string{}),
		msgCh:        msgCh,
		msgCount:     tui.NewState(0),
		feed:         tui.NewRef(),
		scrollY:      tui.NewState(0),
	}

	go func() {
		msgs := []string{
			"Hello from producer",
			"Data packet received",
			"Processing complete",
			"Checkpoint saved",
			"Heartbeat OK",
		}
		i := 0
		for {
			time.Sleep(3 * time.Second)
			msgCh <- msgs[i%len(msgs)]
			i++
		}
	}()

	return app
}

func (w *watcherApp) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnRune('q', func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnRune('s', func(ke tui.KeyEvent) {
			w.stopwatchOn.Set(!w.stopwatchOn.Get())
		}),
		tui.OnRune('r', func(ke tui.KeyEvent) {
			w.stopwatchSec.Set(0)
			w.stopwatchOn.Set(false)
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
		w.stopwatchSec.Update(func(v int) int { return v + 1 })
	}
}

func (w *watcherApp) addMessage(msg string) {
	w.msgCount.Update(func(v int) int { return v + 1 })
	ts := time.Now().Format("15:04:05")
	entry := fmt.Sprintf("[%s] #%d: %s", ts, w.msgCount.Get(), msg)
	w.messages.Set(append(w.messages.Get(), entry))

	// Auto-scroll to bottom
	el := w.feed.El()
	if el != nil {
		_, maxY := el.MaxScroll()
		w.scrollY.Set(maxY + 1)
	}
}

func (w *watcherApp) scrollBy(delta int) {
	el := w.feed.El()
	if el == nil {
		return
	}
	_, maxY := el.MaxScroll()
	newY := w.scrollY.Get() + delta
	if newY < 0 {
		newY = 0
	}
	if newY > maxY {
		newY = maxY
	}
	w.scrollY.Set(newY)
}

func (w *watcherApp) HandleMouse(me tui.MouseEvent) bool {
	switch me.Button {
	case tui.MouseWheelUp:
		w.scrollBy(-3)
		return true
	case tui.MouseWheelDown:
		w.scrollBy(3)
		return true
	}
	return false
}

func formatDuration(seconds int) string {
	m := seconds / 60
	s := seconds - m*60
	return fmt.Sprintf("%02d:%02d", m, s)
}

templ (w *watcherApp) Render() {
	<div class="flex-col p-1 gap-1 border-rounded border-cyan">
		<span class="text-gradient-cyan-magenta font-bold">Timers & Watchers</span>

		<div class="flex gap-2 shrink-0">
			<span class="font-bold">Stopwatch</span>
			<span class="text-cyan font-bold">{formatDuration(w.stopwatchSec.Get())}</span>
			@if w.stopwatchOn.Get() {
				<span class="text-green font-bold">Running</span>
			} @else {
				<span class="text-yellow">Paused</span>
			}
			<span class="font-dim">[s] toggle  [r] reset</span>
		</div>

		<div class="flex-col border-rounded p-1 grow overflow-y-scroll scrollbar-cyan scrollbar-thumb-bright-cyan"
			ref={w.feed} scrollOffset={0, w.scrollY.Get()}>
			<div class="flex gap-2">
				<span class="font-bold">Live Feed</span>
				<span class="font-dim">{fmt.Sprintf("(%d received)", w.msgCount.Get())}</span>
			</div>
			@for _, msg := range w.messages.Get() {
				<span class="text-green">{msg}</span>
			}
			@if len(w.messages.Get()) == 0 {
				<span class="font-dim">Waiting for messages...</span>
			}
		</div>

		<span class="font-dim">s stopwatch | r reset | q quit</span>
	</div>
}
