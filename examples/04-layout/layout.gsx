package main

import (
	"fmt"
	tui "github.com/grindlemire/go-tui"
)

type alignMode struct {
	name    string
	content tui.AlignContent
}

func alignModes() []alignMode {
	return []alignMode{
		{"content-start", tui.ContentStart},
		{"content-end", tui.ContentEnd},
		{"content-center", tui.ContentCenter},
		{"content-stretch", tui.ContentStretch},
		{"content-between", tui.ContentSpaceBetween},
		{"content-around", tui.ContentSpaceAround},
	}
}

func viewNames() []string {
	return []string{"Dashboard", "Sidebar", "Centered Card", "Flex Wrap"}
}

type layoutApp struct {
	viewIndex *tui.State[int]
	modeIndex *tui.State[int]
}

func LayoutApp() *layoutApp {
	return &layoutApp{
		viewIndex: tui.NewState(0),
		modeIndex: tui.NewState(0),
	}
}

func (l *layoutApp) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnRune('q', func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnKey(tui.KeyTab, func(ke tui.KeyEvent) {
			l.viewIndex.Update(func(v int) int { return (v + 1) % len(viewNames()) })
		}),
		tui.OnKeyMod(tui.KeyTab, tui.ModShift, func(ke tui.KeyEvent) {
			l.viewIndex.Update(func(v int) int { return (v - 1 + len(viewNames())) % len(viewNames()) })
		}),
		tui.OnKey(tui.KeyRight, func(ke tui.KeyEvent) {
			l.modeIndex.Update(func(v int) int { return (v + 1) % len(alignModes()) })
		}),
		tui.OnKey(tui.KeyLeft, func(ke tui.KeyEvent) {
			l.modeIndex.Update(func(v int) int { return (v - 1 + len(alignModes())) % len(alignModes()) })
		}),
	}
}

// Sidebar and main content
templ SidebarLayout() {
	<div class="flex h-full">
		<div class="w-20 border-single flex-col p-1">
			<span class="font-bold">Sidebar</span>
			<span>Navigation</span>
			<span>Settings</span>
		</div>
		<div class="grow flex-col p-1">
			<span class="font-bold">Content</span>
			<span>The main area fills the remaining width.</span>
		</div>
	</div>
}

// Centered card
templ CenteredCard() {
	<div class="flex items-center justify-center h-full">
		<div class="border-rounded p-2 flex-col gap-1 w-40">
			<span class="font-bold text-cyan">Welcome</span>
			<hr />
			<span>This card is centered both horizontally and vertically.</span>
		</div>
	</div>
}

// Dashboard grid
templ Dashboard() {
	<div class="flex-col h-full gap-1 p-1">
		<div class="flex gap-1 grow">
			<div class="grow border-rounded p-1 flex-col">
				<span class="font-bold text-cyan">CPU</span>
				<span>45%</span>
			</div>
			<div class="grow border-rounded p-1 flex-col">
				<span class="font-bold text-green">Memory</span>
				<span>2.1 GB</span>
			</div>
			<div class="grow border-rounded p-1 flex-col">
				<span class="font-bold text-yellow">Disk</span>
				<span>67%</span>
			</div>
		</div>
		<div class="flex gap-1 grow">
			<div class="w-2/3 border-rounded p-1 flex-col">
				<span class="font-bold">Network Activity</span>
				<span>Sparkline goes here</span>
			</div>
			<div class="grow border-rounded p-1 flex-col">
				<span class="font-bold">Events</span>
				<span>Log feed goes here</span>
			</div>
		</div>
	</div>
}

func wrapLabels() []string {
	return []string{"Alpha", "Bravo", "Charlie", "Delta", "Echo", "Foxtrot", "Golf", "Hotel"}
}

// Flex-wrap: items wrap to new lines when they overflow
templ FlexWrapGrid(mode alignMode) {
	<div class="flex-col h-full p-1">
		<div class="flex flex-wrap gap-1 grow" alignContent={mode.content}>
			@for _, label := range wrapLabels() {
				<div class="border-rounded p-1 w-16 flex-col items-center shrink-0">
					<span>{label}</span>
				</div>
			}
		</div>
		<div class="p-1">
			<span class="font-dim">{fmt.Sprintf("align-content: %s  (←/→ to cycle)", mode.name)}</span>
		</div>
	</div>
}

templ ViewHeader(viewIndex int) {
	<div class="flex gap-1 p-1 border-single">
		@for i, name := range viewNames() {
			@if i == viewIndex {
				<span class="font-bold text-cyan">{fmt.Sprintf("[%s]", name)}</span>
			} @else {
				<span class="font-dim">{name}</span>
			}
		}
	</div>
}

templ ViewFooter() {
	<div class="p-1 border-single">
		<span class="font-dim">{"tab/shift+tab: switch view | q: quit"}</span>
	</div>
}

templ (l *layoutApp) Render() {
	<div class="flex-col h-full w-full" deps={l.viewIndex, l.modeIndex}>
		@ViewHeader(l.viewIndex.Get())
		<div class="grow flex-col">
			@if l.viewIndex.Get() == 0 {
				@Dashboard()
			} @else @if l.viewIndex.Get() == 1 {
				@SidebarLayout()
			} @else @if l.viewIndex.Get() == 2 {
				@CenteredCard()
			} @else {
				@FlexWrapGrid(alignModes()[l.modeIndex.Get()])
			}
		</div>
		@ViewFooter()
	</div>
}
