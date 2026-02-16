package main

import (
	"fmt"
	"math"
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
	s.stickToBottom.Set(newY >= maxY)
}

func (s *streamingApp) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnRune('q', func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnRune('j', func(ke tui.KeyEvent) { s.scrollBy(1) }),
		tui.OnRune('k', func(ke tui.KeyEvent) { s.scrollBy(-1) }),
		tui.OnKey(tui.KeyUp, func(ke tui.KeyEvent) { s.scrollBy(-1) }),
		tui.OnKey(tui.KeyDown, func(ke tui.KeyEvent) { s.scrollBy(1) }),
		tui.OnKey(tui.KeyPageUp, func(ke tui.KeyEvent) { s.scrollBy(-10) }),
		tui.OnKey(tui.KeyPageDown, func(ke tui.KeyEvent) { s.scrollBy(10) }),
		tui.OnKey(tui.KeyHome, func(ke tui.KeyEvent) {
			s.scrollY.Set(0)
			s.stickToBottom.Set(false)
		}),
		tui.OnKey(tui.KeyEnd, func(ke tui.KeyEvent) {
			s.scrollY.Set(math.MaxInt)
			s.stickToBottom.Set(true)
		}),
		tui.OnRune(' ', func(ke tui.KeyEvent) {
			if s.stickToBottom.Get() {
				s.stickToBottom.Set(false)
			} else {
				s.scrollY.Set(math.MaxInt)
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
	current := s.lines.Get()
	s.lines.Set(append(current, line))
	if s.stickToBottom.Get() {
		s.scrollY.Set(math.MaxInt)
	}
}

func lineColor(line string) string {
	if len(line) < 20 {
		return ""
	}
	// Color based on metric type
	for i := 0; i < len(line)-3; i++ {
		sub := line[i : i+3]
		if sub == "cpu" {
			return "text-cyan"
		}
		if sub == "mem" {
			return "text-magenta"
		}
		if sub == "net" {
			return "text-green"
		}
		if sub == "dis" {
			return "text-yellow"
		}
		if sub == "io:" {
			return "text-blue"
		}
	}
	return ""
}

templ (s *streamingApp) Render() {
	<div class="flex-col gap-1 p-1 h-full border-rounded border-cyan">
		<div class="flex justify-between shrink-0">
			<span class="text-gradient-cyan-magenta font-bold shrink-0">Live Stream</span>
			<span class="text-cyan font-bold" minWidth={0}>{fmt.Sprintf("%d lines", len(s.lines.Get()))}</span>
		</div>
		<div
			ref={s.content}
			class="flex-col flex-grow border-single p-1"
			scrollable={tui.ScrollVertical}
			scrollOffset={0, s.scrollY.Get()}
		>
			@for _, line := range s.lines.Get() {
				<span class={lineColor(line)}>{line}</span>
			}
		</div>

		<div class="flex gap-2 shrink-0 justify-center">
			<span class="font-dim">Elapsed:</span>
			<span class="text-cyan font-bold">{fmt.Sprintf("%ds", s.elapsed.Get())}</span>
			<span class="font-dim">Auto-scroll:</span>
			@if s.stickToBottom.Get() {
				<span class="text-green font-bold">ON</span>
			} @else {
				<span class="text-yellow">OFF</span>
			}
		</div>

		<span class="font-dim shrink-0">j/k scroll | Space toggle auto-scroll | q quit</span>
	</div>
}
