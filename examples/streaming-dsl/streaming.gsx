package main

import (
	"fmt"
	"time"
	tui "github.com/grindlemire/go-tui"
)

// StreamApp - fully declarative with onChannel and onTimer in DSL
// content ref is declared explicitly for cross-element access
// Scrolling uses built-in arrow keys and mouse wheel (Element.handleScrollEvent)
templ StreamApp(dataCh <-chan string) {
	lineCount := tui.NewState(0)
	elapsed := tui.NewState(0)
	content := tui.NewRef()
	<div
		class="flex-col"
		onTimer={tui.OnTimer(time.Second, tickElapsed(elapsed))}
		onChannel={tui.Watch(dataCh, addLine(lineCount, content))}>
		// Header
		<div
			class="border-blue"
			border={tui.BorderSingle}
			height={3}
			direction={tui.Row}
			justify={tui.JustifyCenter}
			align={tui.AlignCenter}>
			<span class="font-bold text-white">{"Streaming DSL Demo - Use arrow keys to scroll, q to quit"}</span>
		</div>

		// Content area with ref - scrolling handled by Element.handleScrollEvent
		<div
			ref={content}
			class="flex-col border-cyan"
			border={tui.BorderSingle}
			flexGrow={1}
			scrollable={tui.ScrollVertical}
			focusable={true}></div>

		// Footer with reactive state (auto-updates when lineCount/elapsed change)
		<div
			class="border-blue"
			border={tui.BorderSingle}
			height={3}
			direction={tui.Row}
			justify={tui.JustifyCenter}
			align={tui.AlignCenter}>
			<span class="text-white">
				{fmt.Sprintf("Lines: %d | Elapsed: %ds | Press q to exit", lineCount.Get(), elapsed.Get())}
			</span>
		</div>
	</div>
}

func tickElapsed(elapsed *tui.State[int]) func() {
	return func() {
		elapsed.Set(elapsed.Get() + 1)
	}
}

func addLine(lineCount *tui.State[int], content *tui.Ref) func(string) {
	return func(line string) {
		lineCount.Set(lineCount.Get() + 1)

		el := content.El()
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
}
