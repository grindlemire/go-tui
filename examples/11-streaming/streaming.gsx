package main

import (
	"fmt"
	"time"
	"github.com/grindlemire/go-tui/pkg/tui"
	"github.com/grindlemire/go-tui/pkg/tui/element"
)

templ Streaming(dataCh <-chan string) {
	lineCount := tui.NewState(0)
	elapsed := tui.NewState(0)
	<div class="flex-col gap-1 p-1"
	     onTimer={tui.OnTimer(time.Second, tick(elapsed))}
	     onChannel={tui.Watch(dataCh, addLine(lineCount, Content))}>
		<span class="font-bold text-cyan">Streaming with Channels and Timers</span>
		<hr class="border" />

		<div #Content
		     class="border-single p-1 flex-col flex-grow overflow-y-scroll"
		     focusable={true}
		     onKeyPress={handleScrollKeys(Content)}></div>

		<div class="flex gap-2">
			<span>Lines: {fmt.Sprintf("%d", lineCount.Get())}</span>
			<span>Elapsed: {fmt.Sprintf("%ds", elapsed.Get())}</span>
		</div>

		<span class="font-dim">Press q to quit</span>
	</div>
}

func tick(elapsed *tui.State[int]) func() {
	return func() {
		elapsed.Set(elapsed.Get() + 1)
	}
}

func addLine(lineCount *tui.State[int], content *element.Element) func(string) {
	return func(line string) {
		lineCount.Set(lineCount.Get() + 1)

		stayAtBottom := content.IsAtBottom()

		lineElem := element.New(
			element.WithText(line),
			element.WithTextStyle(tui.NewStyle().Foreground(tui.Green)),
		)
		content.AddChild(lineElem)

		if stayAtBottom {
			content.ScrollToBottom()
		}
	}
}

func handleScrollKeys(content *element.Element) func(tui.KeyEvent) {
	return func(e tui.KeyEvent) {
		switch e.Rune {
		case 'j':
			content.ScrollBy(0, 1)
		case 'k':
			content.ScrollBy(0, -1)
		}
	}
}
