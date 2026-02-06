package main

import (
	"fmt"
	"time"
	tui "github.com/grindlemire/go-tui"
)

type streamingApp struct {
	dataCh    <-chan string
	lineCount *tui.State[int]
	elapsed   *tui.State[int]
	content   *tui.Ref
}

func Streaming(dataCh <-chan string) *streamingApp {
	return &streamingApp{
		dataCh:    dataCh,
		lineCount: tui.NewState(0),
		elapsed:   tui.NewState(0),
		content:   tui.NewRef(),
	}
}

func (s *streamingApp) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnRune('q', func(ke tui.KeyEvent) { tui.Stop() }),
		tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { tui.Stop() }),
		// Custom j/k scrolling
		tui.OnRune('j', func(ke tui.KeyEvent) {
			if s.content.El() != nil {
				s.content.El().ScrollBy(0, 1)
			}
		}),
		tui.OnRune('k', func(ke tui.KeyEvent) {
			if s.content.El() != nil {
				s.content.El().ScrollBy(0, -1)
			}
		}),
	}
}

func (s *streamingApp) tick() {
	s.elapsed.Set(s.elapsed.Get() + 1)
}

func (s *streamingApp) addLine(line string) {
	s.lineCount.Set(s.lineCount.Get() + 1)

	el := s.content.El()
	stayAtBottom := el.IsAtBottom()

	lineElem := tui.New(
		tui.WithText(line),
		tui.WithTextStyle(tui.NewStyle().Foreground(tui.Green)),
	)
	el.AddChild(lineElem)

	if stayAtBottom {
		el.ScrollToBottom()
	}
}

templ (s *streamingApp) Render() {
	content := s.content
	<div class="flex-col gap-1 p-1"
	     onTimer={tui.OnTimer(time.Second, s.tick)}
	     onChannel={tui.Watch(s.dataCh, s.addLine)}>
		<span class="font-bold text-cyan">Streaming with Channels and Timers</span>
		<hr class="border" />

		<div ref={content}
		     class="border-single p-1 flex-col flex-grow overflow-y-scroll"
		     focusable={true}></div>

		<div class="flex gap-2">
			<span>Lines: {fmt.Sprintf("%d", s.lineCount.Get())}</span>
			<span>Elapsed: {fmt.Sprintf("%ds", s.elapsed.Get())}</span>
		</div>

		<span class="font-dim">Press q to quit, j/k to scroll</span>
	</div>
}
