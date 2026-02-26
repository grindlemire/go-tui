package main

import (
	"fmt"
	tui "github.com/grindlemire/go-tui"
)

type elementsApp struct {
	progress *tui.State[int]
	scrollY  *tui.State[int]
	content  *tui.Ref
}

func Elements() *elementsApp {
	return &elementsApp{
		progress: tui.NewState(62),
		scrollY:  tui.NewState(0),
		content:  tui.NewRef(),
	}
}

func (e *elementsApp) scrollBy(delta int) {
	el := e.content.El()
	if el == nil {
		return
	}
	_, maxY := el.MaxScroll()
	newY := e.scrollY.Get() + delta
	if newY < 0 {
		newY = 0
	} else if newY > maxY {
		newY = maxY
	}
	e.scrollY.Set(newY)
}

func (e *elementsApp) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnRune('q', func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnRune('+', func(ke tui.KeyEvent) {
			v := e.progress.Get() + 5
			if v > 100 {
				v = 100
			}
			e.progress.Set(v)
		}),
		tui.OnRune('-', func(ke tui.KeyEvent) {
			v := e.progress.Get() - 5
			if v < 0 {
				v = 0
			}
			e.progress.Set(v)
		}),
		tui.OnRune('j', func(ke tui.KeyEvent) { e.scrollBy(1) }),
		tui.OnRune('k', func(ke tui.KeyEvent) { e.scrollBy(-1) }),
		tui.OnKey(tui.KeyDown, func(ke tui.KeyEvent) { e.scrollBy(1) }),
		tui.OnKey(tui.KeyUp, func(ke tui.KeyEvent) { e.scrollBy(-1) }),
	}
}

func (e *elementsApp) HandleMouse(me tui.MouseEvent) bool {
	switch me.Button {
	case tui.MouseWheelUp:
		e.scrollBy(-1)
		return true
	case tui.MouseWheelDown:
		e.scrollBy(1)
		return true
	}
	return false
}

func progressBar(value, width int) string {
	filled := value * width / 100
	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	return bar
}

templ (e *elementsApp) Render() {
	<div
		ref={e.content}
		class="flex-col gap-1 h-full"
		scrollable={tui.ScrollVertical}
		scrollOffset={0, e.scrollY.Get()}
	>
		<span class="text-gradient-cyan-magenta font-bold">Built-in Elements</span>

		// Text Elements
		<div class="flex-col border-rounded p-1 gap-1">
			<span class="text-gradient-cyan-magenta font-bold">Text Elements</span>
			<p>{"Paragraph text (<p>) wraps automatically when the content exceeds the available width. This demonstrates how longer text content is displayed."}</p>
			<hr />
			<span class="text-cyan">{"This is a <span> element for inline styled text"}</span>
			<br />
			<span class="font-dim">{"<hr> above draws a line, <br> inserts a blank line"}</span>
		</div>

		// Lists and Table side by side
		<div class="flex gap-1">
			<div class="flex-col border-rounded p-1 gap-1">
				<span class="text-gradient-cyan-magenta font-bold">{"Lists (<ul> / <li>)"}</span>
				<ul class="flex-col p-1">
					<li>
						<span>First item</span>
					</li>
					<li>
						<span>Second item</span>
					</li>
					<li>
						<span>Third item</span>
					</li>
					<li>
						<span class="text-cyan">Fourth (styled)</span>
					</li>
				</ul>
			</div>
			<div class="flex-col border-rounded p-1 gap-1">
				<span class="text-gradient-cyan-magenta font-bold">Table</span>
				<table class="p-1">
					<tr>
						<th>Name</th>
						<th>Role</th>
						<th>Lvl</th>
					</tr>
					<hr />
					<tr>
						<td class="text-cyan">Alice</td>
						<td>Engineer</td>
						<td class="text-green">Sr</td>
					</tr>
					<tr>
						<td class="text-cyan">Bob</td>
						<td>Designer</td>
						<td class="text-yellow">Jr</td>
					</tr>
				</table>
			</div>
		</div>

		// Buttons
		<div class="flex-col border-rounded p-1 gap-1">
			<span class="text-gradient-cyan-magenta font-bold">Buttons</span>
			<div class="flex gap-2">
				<button>{"Save"}</button>
				<button>{"Cancel"}</button>
				<button class="font-bold">{"Submit"}</button>
				<button disabled={true}>{"Disabled"}</button>
			</div>
		</div>

		// Progress bars (using custom rendering since <progress> attributes aren't supported yet)
		<div class="flex-col border-rounded p-1 gap-1">
			<span class="text-gradient-cyan-magenta font-bold">Progress Bars</span>
			<div class="flex gap-2 items-center">
				<span class="font-dim w-10">Download:</span>
				<span class="text-cyan">{progressBar(e.progress.Get(), 25)}</span>
				<span class="text-cyan font-bold">{fmt.Sprintf("%d%%", e.progress.Get())}</span>
			</div>
			<div class="flex gap-2 items-center">
				<span class="font-dim w-10">Upload:</span>
				<span class="text-green">{progressBar(100, 25)}</span>
				<span class="text-green font-bold">{"100%"}</span>
			</div>
			<div class="flex gap-2 items-center">
				<span class="font-dim w-10">Build:</span>
				<span class="text-yellow">{progressBar(35, 25)}</span>
				<span class="text-yellow font-bold">{"35%"}</span>
			</div>
		</div>

		<span class="font-dim">+/- adjust progress | j/k scroll | q quit</span>
	</div>
}
