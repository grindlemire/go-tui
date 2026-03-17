package main

import (
	"fmt"
	"math"
	tui "github.com/grindlemire/go-tui"
)

type feedApp struct {
	messages      *tui.State[[]string]
	paused        *tui.State[bool]
	scrollY       *tui.State[int]
	stickToBottom *tui.State[bool]
	content       *tui.Ref
	mode          string
}

func NewFeedApp(mode string) *feedApp {
	return &feedApp{
		messages:      tui.NewState([]string{}),
		paused:        tui.NewState(false),
		scrollY:       tui.NewState(0),
		stickToBottom: tui.NewState(false),
		content:       tui.NewRef(),
		mode:          mode,
	}
}

func (f *feedApp) scrollBy(delta int) {
	el := f.content.El()
	if el == nil {
		return
	}
	// Read actual position from element, not from state (state may be math.MaxInt).
	_, curY := el.ScrollOffset()
	_, maxY := el.MaxScroll()
	newY := curY + delta
	if newY < 0 {
		newY = 0
	} else if newY > maxY {
		newY = maxY
	}
	f.scrollY.Set(newY)
	f.stickToBottom.Set(false)
}

func (f *feedApp) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnStop(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnStop(tui.Rune('q'), func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnStop(tui.Rune('p'), func(ke tui.KeyEvent) {
			f.paused.Set(!f.paused.Get())
		}),
		tui.OnStop(tui.Rune('s'), func(ke tui.KeyEvent) {
			if f.stickToBottom.Get() {
				// Turning off: capture actual scroll position so scrollY
				// is no longer math.MaxInt (which would keep following the bottom).
				if el := f.content.El(); el != nil {
					_, y := el.ScrollOffset()
					f.scrollY.Set(y)
				}
				f.stickToBottom.Set(false)
			} else {
				f.stickToBottom.Set(true)
			}
		}),
		tui.On(tui.Rune('j'), func(ke tui.KeyEvent) { f.scrollBy(1) }),
		tui.On(tui.Rune('k'), func(ke tui.KeyEvent) { f.scrollBy(-1) }),
		tui.On(tui.KeyUp, func(ke tui.KeyEvent) { f.scrollBy(-1) }),
		tui.On(tui.KeyDown, func(ke tui.KeyEvent) { f.scrollBy(1) }),
		tui.On(tui.KeyPageUp, func(ke tui.KeyEvent) { f.scrollBy(-10) }),
		tui.On(tui.KeyPageDown, func(ke tui.KeyEvent) { f.scrollBy(10) }),
		tui.On(tui.KeyHome, func(ke tui.KeyEvent) {
			f.scrollY.Set(0)
			f.stickToBottom.Set(false)
		}),
	}
}

func (f *feedApp) HandleMouse(me tui.MouseEvent) bool {
	switch me.Button {
	case tui.MouseWheelUp:
		f.scrollBy(-1)
		return true
	case tui.MouseWheelDown:
		f.scrollBy(1)
		return true
	}
	return false
}

func (f *feedApp) AddMessage(msg string) {
	f.messages.Update(func(msgs []string) []string {
		return append(msgs, msg)
	})
	if f.stickToBottom.Get() {
		f.scrollY.Set(math.MaxInt)
	}
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

func stickyLabel(sticky bool) string {
	if sticky {
		return "STICKY"
	}
	return "FREE"
}

func stickyClass(sticky bool) string {
	if sticky {
		return "text-cyan font-bold"
	}
	return "text-yellow"
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
		<div
			ref={f.content}
			class="flex-col flex-grow border-single p-1"
			scrollable={tui.ScrollVertical}
			scrollOffset={0, f.scrollY.Get()}
		>
			for _, msg := range f.messages.Get() {
				<span class="font-dim">{msg}</span>
			}
		</div>
		<hr />
		<div class="flex justify-between px-1 shrink-0">
			<div class="flex gap-2">
				<span class="font-dim">p: pause</span>
				<span class="font-dim">s: sticky</span>
				<span class="font-dim">j/k: scroll</span>
				<span class="font-dim">q: quit</span>
			</div>
			<div class="flex gap-2">
				<span class={pauseClass(f.paused.Get())}>{pauseLabel(f.paused.Get())}</span>
				<span class={stickyClass(f.stickToBottom.Get())}>{stickyLabel(f.stickToBottom.Get())}</span>
				<span class="font-dim">{fmt.Sprintf("%d msgs", len(f.messages.Get()))}</span>
			</div>
		</div>
	</div>
}
