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
	}
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
	<div class="flex-col gap-1 p-1 h-full border-rounded">
		<span class="text-gradient-cyan-blue font-bold shrink-0">{"Streaming with Channels and Timers"}</span>
		<hr class="border shrink-0" />

		<div ref={s.content}
		     class="border-single p-1 flex-col flex-grow"
		     scrollable={tui.ScrollVertical}
		     focusable={true}></div>

		<div class="flex gap-2 shrink-0 justify-center">
			<span class="font-dim">{"Lines:"}</span>
			<span class="text-cyan font-bold">{fmt.Sprintf("%d", s.lineCount.Get())}</span>
			<span class="font-dim">{"Elapsed:"}</span>
			<span class="text-cyan font-bold">{fmt.Sprintf("%ds", s.elapsed.Get())}</span>
		</div>

		<span class="font-dim shrink-0">{"Arrow keys to scroll | [q] quit"}</span>
	</div>
}
