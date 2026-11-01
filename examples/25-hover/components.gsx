package main

import (
	"fmt"
	tui "github.com/grindlemire/go-tui"
)

// componentsPage shows hover on component tags (<input>, <textarea>) and drag
// reporting.
type componentsPage struct {
	d        *hoverDemo
	name     *tui.State[string]
	dragArea *tui.Ref
	dragging bool
	from     string
	drag     *tui.State[string]
	steps    int
}

func newComponentsPage(d *hoverDemo) *componentsPage {
	return &componentsPage{
		d:        d,
		name:     tui.NewState(""),
		dragArea: tui.NewRef(),
		drag:     tui.NewState("press and drag in here"),
	}
}

func (p *componentsPage) HandleMouse(me tui.MouseEvent) bool {
	pos := fmt.Sprintf("%d,%d", me.X, me.Y)
	switch {
	case me.Action == tui.MousePress && me.Button == tui.MouseLeft:
		if el := p.dragArea.El(); el != nil && el.ContainsPoint(me.X, me.Y) {
			p.dragging, p.from, p.steps = true, pos, 0
			p.drag.Set("pressed at " + pos)
			return true
		}
	case me.Action == tui.MouseDrag && p.dragging:
		p.steps++
		p.drag.Set(fmt.Sprintf("dragging %s -> %s (%d drag events)", p.from, pos, p.steps))
		return true
	case me.Action == tui.MouseRelease && p.dragging:
		p.dragging = false
		p.drag.Set(fmt.Sprintf("dragged %s -> %s (%d drag events)", p.from, pos, p.steps))
		return true
	}
	return false
}

templ (p *componentsPage) Render() {
	<div class="flex gap-2 w-full h-full">
		<div class="flex-col">
			<span class="font-dim">input (Tab to focus, Esc to leave):</span>
			<input
				value={p.name}
				placeholder="hover me"
				width={28}
				border={tui.BorderRounded}
				class="hover:border-cyan"
				onMouseEnter={p.d.logged("enter input")}
				onMouseLeave={p.d.logged("leave input")}
			/>
			<span class="font-dim">textarea:</span>
			<textarea
				placeholder="hover me too"
				width={28}
				maxHeight={3}
				border={tui.BorderRounded}
				class="hover:border-magenta"
				onMouseEnter={p.d.logged("enter textarea")}
				onMouseLeave={p.d.logged("leave textarea")}
			/>
		</div>
		<div
			ref={p.dragArea}
			class="flex-col border-rounded px-1 w-36 justify-center items-center hover:border-yellow"
			height={9}
			borderTitle="Drag area"
			onMouseEnter={p.d.logged("enter drag area")}
			onMouseLeave={p.d.logged("leave drag area")}>
			<span>{p.drag.Get()}</span>
			<span class="font-dim">Hover still tracks while dragging.</span>
		</div>
	</div>
}
