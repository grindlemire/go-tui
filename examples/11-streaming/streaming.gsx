package main

import (
	"fmt"
	"time"
	tui "github.com/grindlemire/go-tui"
)

type streamingApp struct {
	dataCh        <-chan string
	lines         *tui.State[[]string]
	scrollY       *tui.State[int]
	stickToBottom *tui.State[bool]
	elapsed       *tui.State[int]
	content       *tui.Ref
}

func Streaming(dataCh <-chan string) *streamingApp {
	return &streamingApp{
		dataCh:        dataCh,
		lines:         tui.NewState([]string{}),
		scrollY:       tui.NewState(0),
		stickToBottom: tui.NewState(true),
		elapsed:       tui.NewState(0),
		content:       tui.NewRef(),
	}
}

func (s *streamingApp) scrollBy(delta int) {
	el := s.content.El()
	if el == nil {
		return
	}
	_, maxY := el.MaxScroll()
	newY := s.scrollY.Get() + delta
	if newY < 0 {
		newY = 0
	} else if newY > maxY {
		newY = maxY
	}
	s.scrollY.Set(newY)

	// Update stickToBottom based on whether we're at bottom
	s.stickToBottom.Set(newY >= maxY)
}

func (s *streamingApp) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnRune('q', func(ke tui.KeyEvent) { tui.Stop() }),
		tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { tui.Stop() }),
		tui.OnKey(tui.KeyUp, func(ke tui.KeyEvent) { s.scrollBy(-1) }),
		tui.OnKey(tui.KeyDown, func(ke tui.KeyEvent) { s.scrollBy(1) }),
		tui.OnRune('k', func(ke tui.KeyEvent) { s.scrollBy(-1) }),
		tui.OnRune('j', func(ke tui.KeyEvent) { s.scrollBy(1) }),
		tui.OnKey(tui.KeyPageUp, func(ke tui.KeyEvent) { s.scrollBy(-10) }),
		tui.OnKey(tui.KeyPageDown, func(ke tui.KeyEvent) { s.scrollBy(10) }),
		tui.OnKey(tui.KeyHome, func(ke tui.KeyEvent) {
			s.scrollY.Set(0)
			s.stickToBottom.Set(false)
		}),
		tui.OnKey(tui.KeyEnd, func(ke tui.KeyEvent) {
			el := s.content.El()
			if el != nil {
				_, maxY := el.MaxScroll()
				s.scrollY.Set(maxY)
				s.stickToBottom.Set(true)
			}
		}),
	}
}

func (s *streamingApp) HandleMouse(me tui.MouseEvent) bool {
	switch me.Button {
	case tui.MouseWheelUp:
		s.scrollBy(-1)
		return true
	case tui.MouseWheelDown:
		s.scrollBy(1)
		return true
	}
	return false
}

func (s *streamingApp) Watchers() []tui.Watcher {
	return []tui.Watcher{
		tui.OnTimer(time.Second, s.tick),
		tui.Watch(s.dataCh, s.addLine),
	}
}

func (s *streamingApp) tick() {
	s.elapsed.Set(s.elapsed.Get() + 1)
}

func (s *streamingApp) addLine(line string) {
	// Append the line to our state
	current := s.lines.Get()
	s.lines.Set(append(current, line))

	// If stickToBottom, set scrollY to a very large value
	// It will be clamped to maxY during layout
	if s.stickToBottom.Get() {
		s.scrollY.Set(999999)
	}
}

func (s *streamingApp) getScrollY() int {
	// If stickToBottom is true, return a large value that will be clamped
	if s.stickToBottom.Get() {
		return 999999
	}
	return s.scrollY.Get()
}

templ (s *streamingApp) Render() {
	<div class="flex-col gap-1 p-1 h-full border-rounded">
		<span class="text-gradient-cyan-blue font-bold shrink-0">{"Streaming with Channels and Timers"}</span>
		<hr class="border shrink-0" />
		<div
			ref={s.content}
			class="border-single p-1 flex-col flex-grow"
			scrollable={tui.ScrollVertical}
			scrollOffset={0, s.getScrollY()}>
			@for _, line := range s.lines.Get() {
				<span class="text-green">{line}</span>
			}
		</div>

		<div class="flex gap-2 shrink-0 justify-center">
			<span class="font-dim">{"Lines:"}</span>
			<span class="text-cyan font-bold">{fmt.Sprintf("%d", len(s.lines.Get()))}</span>
			<span class="font-dim">{"Elapsed:"}</span>
			<span class="text-cyan font-bold">{fmt.Sprintf("%ds", s.elapsed.Get())}</span>
		</div>

		<span class="font-dim shrink-0">{"↑↓/jk scroll | [q] quit"}</span>
	</div>
}
