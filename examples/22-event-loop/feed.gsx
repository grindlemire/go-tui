package main

import (
	"fmt"
	tui "github.com/grindlemire/go-tui"
)

type feedApp struct {
	messages *tui.State[[]string]
	paused   *tui.State[bool]
	mode     string
}

func NewFeedApp(mode string) *feedApp {
	return &feedApp{
		messages: tui.NewState([]string{}),
		paused:   tui.NewState(false),
		mode:     mode,
	}
}

func (f *feedApp) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnStop(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnStop(tui.Rune('q'), func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnStop(tui.Rune('p'), func(ke tui.KeyEvent) {
			f.paused.Set(!f.paused.Get())
		}),
	}
}

func (f *feedApp) AddMessage(msg string) {
	f.messages.Update(func(msgs []string) []string {
		return append(msgs, msg)
	})
}

func (f *feedApp) IsPaused() bool {
	return f.paused.Get()
}

func pauseLabel(paused bool) string {
	if paused {
		return "PAUSED"
	}
	return "LIVE"
}

func pauseClass(paused bool) string {
	if paused {
		return "text-yellow font-bold"
	}
	return "text-green font-bold"
}

func lastN(msgs []string, n int) []string {
	if len(msgs) <= n {
		return msgs
	}
	return msgs[len(msgs)-n:]
}

templ (f *feedApp) Render() {
	<div class="flex-col h-full border-rounded border-cyan">
		<div class="flex justify-between px-1 shrink-0">
			<span class="text-gradient-cyan-magenta font-bold">Event Loop Demo</span>
			<div class="flex gap-1">
				<span class="font-dim">mode:</span>
				<span class="text-cyan font-bold">{f.mode}</span>
			</div>
		</div>
		<hr />
		<div class="flex-col grow p-1 min-h-0 overflow-y-scroll">
			for _, msg := range lastN(f.messages.Get(), 100) {
				<span class="font-dim">{msg}</span>
			}
		</div>
		<hr />
		<div class="flex justify-between px-1 shrink-0">
			<div class="flex gap-2">
				<span class="font-dim">p: toggle</span>
				<span class="font-dim">q: quit</span>
			</div>
			<div class="flex gap-2">
				<span class={pauseClass(f.paused.Get())}>{pauseLabel(f.paused.Get())}</span>
				<span class="font-dim">{fmt.Sprintf("%d msgs", len(f.messages.Get()))}</span>
			</div>
		</div>
	</div>
}
