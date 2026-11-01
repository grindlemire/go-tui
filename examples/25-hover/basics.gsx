package main

import tui "github.com/grindlemire/go-tui"

var inbox = []string{
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
	"Team offsite agenda",
	"Coffee chat?",
	"Dependency update available",
	"Quarterly OKRs",
}

// basicsPage shows hover: classes on a scrollable list and a card that stays
// lit while the pointer moves between its buttons.
type basicsPage struct {
	d *hoverDemo
}

templ (p *basicsPage) Render() {
	<div class="flex gap-2 w-full h-full">
		<div
			class="flex-col border-rounded w-32 shrink-0 overflow-y-scroll scrollbar-bright-black"
			height={12}
			borderTitle="Inbox (wheel scrolls)">
			for _, row := range inbox {
				<span key={row} class="px-1 hover:bg-blue hover:text-bright-white" onMouseEnter={p.d.logged("enter row " + row)} onMouseLeave={p.d.logged("leave row " + row)}>{row}</span>
			}
		</div>
		<div class="flex-col px-1 border-rounded border-[#666666] hover:border-magenta" height={10} onMouseEnter={p.d.logged("enter card")} onMouseLeave={p.d.logged("leave card")}>
			<span class="font-bold">Nested card</span>
			<span class="font-dim">The border stays lit</span>
			<span class="font-dim">while you move between</span>
			<span class="font-dim">the buttons inside it.</span>
			<div class="flex gap-1 mt-1">
				<span class="px-1 border-single hover:border-yellow hover:text-yellow hover:font-bold" onMouseEnter={p.d.logged("  enter button Save")} onMouseLeave={p.d.logged("  leave button Save")}>Save</span>
				<span class="px-1 border-single hover:border-yellow hover:text-yellow hover:font-bold" onMouseEnter={p.d.logged("  enter button Discard")} onMouseLeave={p.d.logged("  leave button Discard")}>Discard</span>
			</div>
		</div>
	</div>
}
