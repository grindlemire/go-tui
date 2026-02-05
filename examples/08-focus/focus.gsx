package main

import (
	tui "github.com/grindlemire/go-tui"
)

type focusApp struct {
	focused *tui.State[string]
}

func Focus() *focusApp {
	return &focusApp{
		focused: tui.NewState("(none)"),
	}
}

func (f *focusApp) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnRune('q', func(ke tui.KeyEvent) { tui.Stop() }),
		tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { tui.Stop() }),
	}
}

templ (f *focusApp) Render() {
	<div class="flex-col gap-2 p-2">
		<span class="font-bold text-cyan">Focus Navigation</span>
		<hr class="border" />
		<div class="flex gap-2">
			<div
				class="border-single p-2 w-15 h-5 items-center justify-center"
				focusable={true}
				onFocus={onFocusBox("Box A", f.focused)}
				onBlur={onBlurBox(f.focused)}>
				<span class="text-red">Box A</span>
			</div>
			<div
				class="border-single p-2 w-15 h-5 items-center justify-center"
				focusable={true}
				onFocus={onFocusBox("Box B", f.focused)}
				onBlur={onBlurBox(f.focused)}>
				<span class="text-green">Box B</span>
			</div>
			<div
				class="border-single p-2 w-15 h-5 items-center justify-center"
				focusable={true}
				onFocus={onFocusBox("Box C", f.focused)}
				onBlur={onBlurBox(f.focused)}>
				<span
					class="text-blue"
				>
					Box C
				</span>
			</div>
		</div>
		<div class="flex gap-1">
			<span>Focused:</span>
			<span class="font-bold text-yellow">
				{f.focused.Get()}
			</span>
		</div>
		<span class="font-dim">Press Tab/Shift+Tab to navigate, q to quit</span>
	</div>
}

func onFocusBox(name string, focused *tui.State[string]) func(*tui.Element) {
	return func(el *tui.Element) {
		focused.Set(name)
		el.SetBorder(tui.BorderDouble)
	}
}

func onBlurBox(focused *tui.State[string]) func(*tui.Element) {
	return func(el *tui.Element) {
		focused.Set("(none)")
		el.SetBorder(tui.BorderSingle)
	}
}
