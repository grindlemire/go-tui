package main

import (
	"fmt"
	tui "github.com/grindlemire/go-tui"
)

var tabs = []string{"Inbox", "Drafts", "Sent", "Archive"}

var rows = []string{
	"Quarterly report draft",
	"Lunch on Thursday?",
	"Build failed on main",
	"Re: hover support",
	"Your invoice is ready",
	"Standup moved to 10:30",
	"New comment on #178",
	"Weekend plans",
	"Release notes v0.23",
	"Password expires soon",
}

const maxLog = 6

type hoverDemo struct {
	activeTab  *tui.State[string]
	hoveredTab *tui.State[string]
	log        *tui.State[[]string]
	pointer    *tui.State[string]
	moves      *tui.State[int]
	showModal  *tui.State[bool]
	tabRefs    *tui.RefMap[string]
}

func HoverDemo() *hoverDemo {
	return &hoverDemo{
		activeTab:  tui.NewState("Inbox"),
		hoveredTab: tui.NewState(""),
		log:        tui.NewState([]string{}),
		pointer:    tui.NewState("-"),
		moves:      tui.NewState(0),
		showModal:  tui.NewState(false),
		tabRefs:    tui.NewRefMap[string](),
	}
}

func (d *hoverDemo) record(event string) {
	entries := append(d.log.Get(), event)
	if len(entries) > maxLog {
		entries = entries[len(entries)-maxLog:]
	}
	d.log.Set(entries)
}

// logged returns a hover callback that only records the event.
func (d *hoverDemo) logged(event string) func() {
	return func() { d.record(event) }
}

// Tabs drive their hover look from state, to show callbacks.
func (d *hoverDemo) enterTab(tab string) func() {
	return func() {
		d.record("enter tab " + tab)
		d.hoveredTab.Set(tab)
	}
}

func (d *hoverDemo) leaveTab(tab string) func() {
	return func() {
		d.record("leave tab " + tab)
		d.hoveredTab.Set("")
	}
}

func (d *hoverDemo) closeModal() {
	d.showModal.Set(false)
}

func (d *hoverDemo) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.On(tui.Rune('q'), func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.On(tui.Rune('m'), func(ke tui.KeyEvent) { d.showModal.Set(true) }),
		tui.On(tui.Rune('c'), func(ke tui.KeyEvent) { d.log.Set([]string{}) }),
	}
}

func (d *hoverDemo) HandleMouse(me tui.MouseEvent) bool {
	d.pointer.Set(fmt.Sprintf("%d,%d", me.X, me.Y))
	if me.Action == tui.MouseMove {
		d.moves.Set(d.moves.Get() + 1)
	}
	if me.Button == tui.MouseLeft && me.Action == tui.MousePress {
		for name, el := range d.tabRefs.All() {
			if el != nil && el.ContainsPoint(me.X, me.Y) {
				d.activeTab.Set(name)
				return true
			}
		}
	}
	return false
}

func (d *hoverDemo) tabClass(tab string) string {
	switch {
	case tab == d.activeTab.Get():
		return "px-1 font-bold bg-cyan text-black"
	case tab == d.hoveredTab.Get():
		return "px-1 underline text-bright-cyan"
	}
	return "px-1 text-white"
}

templ (d *hoverDemo) Render() {
	<div class="flex-col p-1 gap-1 border-rounded border-cyan">
		<span class="text-gradient-cyan-magenta font-bold">Hover Demo</span>
		<div class="flex">
			for _, tab := range tabs {
				<span ref={d.tabRefs} key={tab} class={d.tabClass(tab)} onMouseEnter={d.enterTab(tab)} onMouseLeave={d.leaveTab(tab)}>{tab}</span>
			}
		</div>
		<div class="flex gap-2">
			<div
				class="flex-col border-rounded w-32 overflow-y-scroll scrollbar-bright-black"
				height={7}
				borderTitle={d.activeTab.Get() + " (scroll me)"}>
				for _, row := range rows {
					<span key={row} class="px-1 hover:bg-blue hover:text-bright-white" onMouseEnter={d.logged("enter row " + row)} onMouseLeave={d.logged("leave row " + row)}>{row}</span>
				}
			</div>
			<div class="flex-col px-1 border-rounded border-[#666666] hover:border-magenta" onMouseEnter={d.logged("enter card")} onMouseLeave={d.logged("leave card")}>
				<span class="font-bold">Nested card</span>
				<span class="font-dim">The border stays lit</span>
				<span class="font-dim">while you move between</span>
				<span class="font-dim">the buttons inside it.</span>
				<div class="flex gap-1 mt-1">
					<span class="px-1 border-single hover:border-yellow hover:text-yellow hover:font-bold" onMouseEnter={d.logged("  enter button Save")} onMouseLeave={d.logged("  leave button Save")}>Save</span>
					<span class="px-1 border-single hover:border-yellow hover:text-yellow hover:font-bold" onMouseEnter={d.logged("  enter button Discard")} onMouseLeave={d.logged("  leave button Discard")}>Discard</span>
				</div>
			</div>
		</div>
		<div class="flex-col border-rounded px-1 h-8" borderTitle="Hover events (newest last)">
			for _, entry := range d.log.Get() {
				<span class="font-dim">{entry}</span>
			}
		</div>
		<span class="font-dim">{fmt.Sprintf("pointer %s | %d move events", d.pointer.Get(), d.moves.Get())}</span>
		<span class="font-dim">click tabs to select | wheel scrolls the list | m modal | c clear log | q quit</span>
		<modal open={d.showModal} class="justify-center items-center">
			<div class="border-rounded p-2 flex-col gap-1 w-40 items-center">
				<span class="font-bold text-yellow">Modal open</span>
				<span class="font-dim">Content behind the backdrop ignores hover.</span>
				<button class="px-2 border-single hover:border-green hover:text-green hover:font-bold" onMouseEnter={d.logged("enter modal button")} onMouseLeave={d.logged("leave modal button")} onActivate={d.closeModal}>Close</button>
			</div>
		</modal>
	</div>
}
