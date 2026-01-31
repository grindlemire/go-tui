package main

import (
	"fmt"
	"time"
	tui "github.com/grindlemire/go-tui"
)

// StreamApp - fully declarative with onChannel and onTimer in DSL
// content ref is declared explicitly for cross-element access
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
			<span class="font-bold text-white">{"Streaming DSL Demo - Use j/k to scroll, q to quit"}</span>
		</div>

		// Content area with ref
		<div
			ref={content}
			class="flex-col border-cyan"
			border={tui.BorderSingle}
			flexGrow={1}
			scrollable={tui.ScrollVertical}
			focusable={true}
			onKeyPress={handleScrollKeys}
			onEvent={handleEvent}></div>

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

func handleScrollKeys(el *tui.Element, e tui.KeyEvent) {
	switch e.Rune {
	case 'j':
		el.ScrollBy(0, 1)
	case 'k':
		el.ScrollBy(0, -1)
	}
}

func handleEvent(el *tui.Element, e tui.Event) bool {
	if mouse, ok := e.(tui.MouseEvent); ok {
		switch mouse.Button {
		case tui.MouseWheelUp:
			el.ScrollBy(0, -1)
			return true
		case tui.MouseWheelDown:
			el.ScrollBy(0, 1)
			return true
		}
	}
	return false
}
