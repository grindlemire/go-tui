package main

import (
	"fmt"

	tui "github.com/grindlemire/go-tui"
)

type panelForm struct {
	panels     []string
	panel1     *tui.State[bool]
	panel2     *tui.State[bool]
	panel3     *tui.State[bool]
	focus      *tui.FocusGroup
	clickCount *tui.State[int]
}

func PanelForm() *panelForm {
	p1 := tui.NewState(true)
	p2 := tui.NewState(false)
	p3 := tui.NewState(false)

	return &panelForm{
		panels:     []string{"Inbox", "Drafts", "Sent"},
		panel1:     p1,
		panel2:     p2,
		panel3:     p3,
		focus:      tui.MustNewFocusGroup(p1, p2, p3),
		clickCount: tui.NewState(0),
	}
}

func (p *panelForm) KeyMap() tui.KeyMap {
	return append(p.focus.KeyMap(), []tui.KeyBinding{
		tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnRune(' ', func(ke tui.KeyEvent) {
			p.clickCount.Update(func(v int) int { return v + 1 })
		}),
	}...)
}

templ (p *panelForm) Render() {
	<div class="flex-col gap-1 p-1">
		<span class="font-bold text-gradient-cyan-magenta">Focus Demo — Tab to switch, Space to interact</span>
		<div class="flex gap-1">
			@for i, name := range p.panels {
				@if i == p.focus.Current() {
					<div class="flex-col border-rounded border-cyan p-1" width={20}>
						<span class="text-cyan font-bold">{name}</span>
						<span class="text-bright-white">{fmt.Sprintf("Actions: %d", p.clickCount.Get())}</span>
					</div>
				} @else {
					<div class="flex-col border-rounded border-black p-1" width={20}>
						<span class="font-dim">{name}</span>
					</div>
				}
			}
		</div>
		<span class="font-dim">Press Esc to quit</span>
	</div>
}
