package main

import (
	"time"
	tui "github.com/grindlemire/go-tui"
)

var gradient = tui.NewGradient(tui.BrightCyan, tui.BrightMagenta)

// phrases to cycle through when the user presses Enter.
var phrases = []string{
	"The quick brown fox jumps over the lazy dog.",
	"Stars scattered across the midnight canvas, each one a whisper of ancient light.",
	"Line one of a multi-line message.\nLine two continues here.\nAnd line three wraps it up.",
	"Streaming text appears character by character, just like a real-time API response.",
}

type streamDemo struct {
	app       *tui.App
	phraseIdx int
	streaming *tui.State[bool]
}

func StreamDemo() *streamDemo {
	return &streamDemo{
		streaming: tui.NewState(false),
	}
}

// streamPhrase opens a StreamAbove writer, writes each character with a
// gradient color and a small delay, then closes the writer.
func (s *streamDemo) streamPhrase() {
	if s.streaming.Get() {
		return
	}
	s.streaming.Set(true)

	text := phrases[s.phraseIdx%len(phrases)]
	s.phraseIdx++

	go func() {
		w := s.app.StreamAbove()
		for _, r := range text {
			w.WriteGradient(string(r), gradient)
			time.Sleep(30 * time.Millisecond)
		}
		w.Close()
		s.app.QueueUpdate(func() {
			s.streaming.Set(false)
		})
	}()
}

func (s *streamDemo) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnKeyStop(tui.KeyEnter, func(ke tui.KeyEvent) { s.streamPhrase() }),
		tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnKey(tui.KeyCtrlC, func(ke tui.KeyEvent) { ke.App().Stop() }),
	}
}

func (s *streamDemo) statusText() string {
	if s.streaming.Get() {
		return "streaming..."
	}
	return "Press Enter to stream a phrase  |  Esc to quit"
}

templ (s *streamDemo) Render() {
	<div class="border-rounded border-cyan items-center justify-center">
		<span class="text-cyan">{s.statusText()}</span>
	</div>
}
