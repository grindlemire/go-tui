package main

import tui "github.com/grindlemire/go-tui"

type layoutApp struct{}

func LayoutApp() *layoutApp {
	return &layoutApp{}
}

func (l *layoutApp) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnRune('q', func(ke tui.KeyEvent) { ke.App().Stop() }),
	}
}

// Full-screen app with header, content, and footer
templ AppLayout() {
	<div class="flex-col h-full">
		<div class="border-single p-1">
			<span class="font-bold text-cyan">My App</span>
		</div>
		<div class="flex-col grow p-1">
			<span>Main content goes here.</span>
			<span>This section grows to fill available space.</span>
		</div>
		<div class="border-single p-1">
			<span class="font-dim">Press q to quit</span>
		</div>
	</div>
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

templ (l *layoutApp) Render() {
	<div class="flex-col h-full w-full">
		@Dashboard()
	</div>
}
