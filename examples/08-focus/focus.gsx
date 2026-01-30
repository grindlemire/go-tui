package main

import (
	"github.com/grindlemire/go-tui/pkg/tui"
	"github.com/grindlemire/go-tui/pkg/tui/element"
)

templ Focus() {
	focused := tui.NewState("(none)")
	<div class="flex-col gap-2 p-2">
		<span class="font-bold text-cyan">Focus Navigation</span>
		<hr class="border" />

		<div class="flex gap-2">
			<div #BoxA class="border-single p-2 w-15 h-5 items-center justify-center"
			     focusable={true}
			     onFocus={onFocusBox("Box A", focused, BoxA)}
			     onBlur={onBlurBox(focused, BoxA)}>
				<span class="text-red">Box A</span>
			</div>
			<div #BoxB class="border-single p-2 w-15 h-5 items-center justify-center"
			     focusable={true}
			     onFocus={onFocusBox("Box B", focused, BoxB)}
			     onBlur={onBlurBox(focused, BoxB)}>
				<span class="text-green">Box B</span>
			</div>
			<div #BoxC class="border-single p-2 w-15 h-5 items-center justify-center"
			     focusable={true}
			     onFocus={onFocusBox("Box C", focused, BoxC)}
			     onBlur={onBlurBox(focused, BoxC)}>
				<span class="text-blue">Box C</span>
			</div>
		</div>

		<div class="flex gap-1">
			<span>Focused:</span>
			<span class="font-bold text-yellow">{focused.Get()}</span>
		</div>

		<span class="font-dim">Press Tab/Shift+Tab to navigate, q to quit</span>
	</div>
}

func onFocusBox(name string, focused *tui.State[string], box *element.Element) func() {
	return func() {
		focused.Set(name)
		// Change border to double when focused
		box.SetBorder(tui.BorderDouble)
	}
}

func onBlurBox(focused *tui.State[string], box *element.Element) func() {
	return func() {
		focused.Set("(none)")
		// Change border back to single when blurred
		box.SetBorder(tui.BorderSingle)
	}
}
