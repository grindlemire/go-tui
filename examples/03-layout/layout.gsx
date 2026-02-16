package main

import tui "github.com/grindlemire/go-tui"

type layoutApp struct {
	scrollY *tui.State[int]
	content *tui.Ref
}

func Layout() *layoutApp {
	return &layoutApp{
		scrollY: tui.NewState(0),
		content: tui.NewRef(),
	}
}

func (l *layoutApp) scrollBy(delta int) {
	el := l.content.El()
	if el == nil {
		return
	}
	_, maxY := el.MaxScroll()
	newY := l.scrollY.Get() + delta
	if newY < 0 {
		newY = 0
	} else if newY > maxY {
		newY = maxY
	}
	l.scrollY.Set(newY)
}

func (l *layoutApp) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnRune('q', func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnRune('j', func(ke tui.KeyEvent) { l.scrollBy(1) }),
		tui.OnRune('k', func(ke tui.KeyEvent) { l.scrollBy(-1) }),
		tui.OnKey(tui.KeyDown, func(ke tui.KeyEvent) { l.scrollBy(1) }),
		tui.OnKey(tui.KeyUp, func(ke tui.KeyEvent) { l.scrollBy(-1) }),
	}
}

func (l *layoutApp) HandleMouse(me tui.MouseEvent) bool {
	switch me.Button {
	case tui.MouseWheelUp:
		l.scrollBy(-1)
		return true
	case tui.MouseWheelDown:
		l.scrollBy(1)
		return true
	}
	return false
}

templ (l *layoutApp) Render() {
	<div
		ref={l.content}
		class="flex-col gap-1 p-1 h-full"
		scrollable={tui.ScrollVertical}
		scrollOffset={0, l.scrollY.Get()}
	>
		<span class="text-gradient-cyan-magenta font-bold">Flexbox Layout</span>

		// Direction
		<div class="flex-col border-rounded p-1 gap-1">
			<span class="text-gradient-cyan-magenta font-bold">Direction</span>
			<div class="flex gap-4">
				<div class="flex-col gap-1">
					<span class="font-dim">flex (row)</span>
					<div class="flex border-single">
						<span class="bg-blue px-2">A</span>
						<span class="bg-green px-2">B</span>
						<span class="bg-magenta px-2">C</span>
					</div>
				</div>
				<div class="flex-col gap-1">
					<span class="font-dim">flex-col (column)</span>
					<div class="flex-col border-single">
						<span class="bg-blue px-2">A</span>
						<span class="bg-green px-2">B</span>
						<span class="bg-magenta px-2">C</span>
					</div>
				</div>
			</div>
		</div>

		// Justify Content
		<div class="flex-col border-rounded p-1 gap-1">
			<span class="text-gradient-cyan-magenta font-bold">Justify Content (main axis)</span>
			<div class="flex-col">
				<span class="font-dim">justify-start</span>
				<div class="flex justify-start w-50 border-single">
					<span class="bg-blue px-1">A</span>
					<span class="bg-green px-1">B</span>
					<span class="bg-magenta px-1">C</span>
				</div>
			</div>
			<div class="flex-col">
				<span class="font-dim">justify-center</span>
				<div class="flex justify-center w-50 border-single">
					<span class="bg-blue px-1">A</span>
					<span class="bg-green px-1">B</span>
					<span class="bg-magenta px-1">C</span>
				</div>
			</div>
			<div class="flex-col">
				<span class="font-dim">justify-end</span>
				<div class="flex justify-end w-50 border-single">
					<span class="bg-blue px-1">A</span>
					<span class="bg-green px-1">B</span>
					<span class="bg-magenta px-1">C</span>
				</div>
			</div>
			<div class="flex-col">
				<span class="font-dim">justify-between</span>
				<div class="flex justify-between w-50 border-single">
					<span class="bg-blue px-1">A</span>
					<span class="bg-green px-1">B</span>
					<span class="bg-magenta px-1">C</span>
				</div>
			</div>
			<div class="flex-col">
				<span class="font-dim">justify-around</span>
				<div class="flex justify-around w-50 border-single">
					<span class="bg-blue px-1">A</span>
					<span class="bg-green px-1">B</span>
					<span class="bg-magenta px-1">C</span>
				</div>
			</div>
			<div class="flex-col">
				<span class="font-dim">justify-evenly</span>
				<div class="flex justify-evenly w-50 border-single">
					<span class="bg-blue px-1">A</span>
					<span class="bg-green px-1">B</span>
					<span class="bg-magenta px-1">C</span>
				</div>
			</div>
		</div>

		// Align Items
		<div class="flex-col border-rounded p-1 gap-1">
			<span class="text-gradient-cyan-magenta font-bold">Align Items (cross axis)</span>
			<div class="flex-col">
				<span class="font-dim">items-start</span>
				<div class="flex-col items-start w-50 border-single">
					<span class="bg-blue px-1">A</span>
					<span class="bg-green px-1">BB</span>
					<span class="bg-magenta px-1">CCC</span>
				</div>
			</div>
			<div class="flex-col">
				<span class="font-dim">items-center</span>
				<div class="flex-col items-center w-50 border-single">
					<span class="bg-blue px-1">A</span>
					<span class="bg-green px-1">BB</span>
					<span class="bg-magenta px-1">CCC</span>
				</div>
			</div>
			<div class="flex-col">
				<span class="font-dim">items-end</span>
				<div class="flex-col items-end w-50 border-single">
					<span class="bg-blue px-1">A</span>
					<span class="bg-green px-1">BB</span>
					<span class="bg-magenta px-1">CCC</span>
				</div>
			</div>
			<div class="flex-col">
				<span class="font-dim">items-stretch</span>
				<div class="flex-col items-stretch w-50 border-single">
					<span class="bg-blue px-1">A</span>
					<span class="bg-green px-1">BB</span>
					<span class="bg-magenta px-1">CCC</span>
				</div>
			</div>
		</div>

		// Gap
		<div class="flex-col border-rounded p-1 gap-1">
			<span class="text-gradient-cyan-magenta font-bold">Gap</span>
			<div class="flex-col">
				<span class="font-dim">gap-0</span>
				<div class="flex w-50 border-single">
					<span class="bg-blue px-1">A</span>
					<span class="bg-green px-1">B</span>
					<span class="bg-magenta px-1">C</span>
				</div>
			</div>
			<div class="flex-col">
				<span class="font-dim">gap-1</span>
				<div class="flex gap-1 w-50 border-single">
					<span class="bg-blue px-1">A</span>
					<span class="bg-green px-1">B</span>
					<span class="bg-magenta px-1">C</span>
				</div>
			</div>
			<div class="flex-col">
				<span class="font-dim">gap-2</span>
				<div class="flex gap-2 w-50 border-single">
					<span class="bg-blue px-1">A</span>
					<span class="bg-green px-1">B</span>
					<span class="bg-magenta px-1">C</span>
				</div>
			</div>
			<div class="flex-col">
				<span class="font-dim">gap-4</span>
				<div class="flex gap-4 w-50 border-single">
					<span class="bg-blue px-1">A</span>
					<span class="bg-green px-1">B</span>
					<span class="bg-magenta px-1">C</span>
				</div>
			</div>
		</div>

		// Flex Grow
		<div class="flex-col border-rounded p-1 gap-1">
			<span class="text-gradient-cyan-magenta font-bold">Flex Grow</span>
			<div class="flex-col">
				<span class="font-dim">No grow</span>
				<div class="flex w-50 border-single">
					<span class="bg-blue px-1">A</span>
					<span class="bg-green px-1">B</span>
					<span class="bg-magenta px-1">C</span>
				</div>
			</div>
			<div class="flex-col">
				<span class="font-dim">B: flex-grow</span>
				<div class="flex w-50 border-single">
					<span class="bg-blue px-1">A</span>
					<span class="bg-green px-1 flex-grow">B</span>
					<span class="bg-magenta px-1">C</span>
				</div>
			</div>
			<div class="flex-col">
				<span class="font-dim">All: flex-grow</span>
				<div class="flex w-50 border-single">
					<span class="bg-blue px-1 flex-grow">A</span>
					<span class="bg-green px-1 flex-grow">B</span>
					<span class="bg-magenta px-1 flex-grow">C</span>
				</div>
			</div>
		</div>

		// Padding
		<div class="flex-col border-rounded p-1 gap-1">
			<span class="text-gradient-cyan-magenta font-bold">Padding</span>
			<div class="flex gap-2">
				<div class="flex-col gap-1">
					<span class="font-dim">p-0</span>
					<div class="border-single bg-yellow">
						<span class="bg-blue text-white">Text</span>
					</div>
				</div>
				<div class="flex-col gap-1">
					<span class="font-dim">p-1</span>
					<div class="border-single bg-yellow p-1">
						<span class="bg-blue text-white">Text</span>
					</div>
				</div>
				<div class="flex-col gap-1">
					<span class="font-dim">p-2</span>
					<div class="border-single bg-yellow p-2">
						<span class="bg-blue text-white">Text</span>
					</div>
				</div>
				<div class="flex-col gap-1">
					<span class="font-dim">px-3 py-1</span>
					<div class="border-single bg-yellow px-3 py-1">
						<span class="bg-blue text-white">Text</span>
					</div>
				</div>
			</div>
		</div>

		// Width Sizing
		<div class="flex-col border-rounded p-1 gap-1">
			<span class="text-gradient-cyan-magenta font-bold">Width Sizing</span>
			<div class="flex-col gap-1 w-60 border-single p-1">
				<div class="flex gap-1">
					<span class="font-dim w-10">w-15</span>
					<div class="bg-blue w-15">
						<span class="text-white">Fixed 15</span>
					</div>
				</div>
				<div class="flex gap-1">
					<span class="font-dim w-10">w-30</span>
					<div class="bg-green w-30">
						<span class="text-white">Fixed 30</span>
					</div>
				</div>
				<div class="flex gap-1">
					<span class="font-dim w-10">w-1/2</span>
					<div class="bg-magenta w-1/2">
						<span class="text-white">Half</span>
					</div>
				</div>
				<div class="flex gap-1">
					<span class="font-dim w-10">w-full</span>
					<div class="bg-cyan w-full">
						<span class="text-black">Full width</span>
					</div>
				</div>
			</div>
		</div>

		// Combination: Holy Grail Layout
		<div class="flex-col border-rounded p-1 gap-1">
			<span class="text-gradient-cyan-magenta font-bold">Combination: Holy Grail Layout</span>
			<div class="flex-col w-60 h-10 border-rounded">
				<div class="flex justify-center bg-blue">
					<span class="font-bold text-white">Header</span>
				</div>
				<div class="flex flex-grow">
					<div class="bg-magenta w-8">
						<span class="text-white px-1">Nav</span>
					</div>
					<div class="flex-grow flex justify-center items-center">
						<span>Content</span>
					</div>
					<div class="bg-cyan w-8">
						<span class="text-black px-1">Side</span>
					</div>
				</div>
				<div class="flex justify-center bg-blue">
					<span class="font-bold text-white">Footer</span>
				</div>
			</div>
		</div>

		// Combination: Dashboard Cards
		<div class="flex-col border-rounded p-1 gap-1">
			<span class="text-gradient-cyan-magenta font-bold">Combination: Dashboard Cards</span>
			<div class="flex gap-1 w-60">
				<div class="flex-col flex-grow border-rounded p-1 gap-1">
					<span class="font-bold text-cyan">Users</span>
					<span class="font-bold">1,234</span>
					<span class="font-dim text-green">{"Up 12%"}</span>
				</div>
				<div class="flex-col flex-grow border-rounded p-1 gap-1">
					<span class="font-bold text-cyan">Revenue</span>
					<span class="font-bold">{"$45k"}</span>
					<span class="font-dim text-red">{"Down 3%"}</span>
				</div>
				<div class="flex-col flex-grow border-rounded p-1 gap-1">
					<span class="font-bold text-cyan">Orders</span>
					<span class="font-bold">567</span>
					<span class="font-dim text-green">{"Up 8%"}</span>
				</div>
			</div>
		</div>

		// Combination: Sidebar Layout
		<div class="flex-col border-rounded p-1 gap-1">
			<span class="text-gradient-cyan-magenta font-bold">Combination: Sidebar Layout</span>
			<div class="flex w-60 h-8 border-rounded">
				<div class="flex-col w-15 border-single gap-1 p-1">
					<span class="font-bold text-cyan">Menu</span>
					<span class="text-green">Home</span>
					<span>About</span>
					<span>Settings</span>
				</div>
				<div class="flex-col flex-grow p-1 gap-1">
					<span class="font-bold">Main Content</span>
					<span class="font-dim">Grows to fill space</span>
				</div>
			</div>
		</div>

		<span class="font-dim">j/k to scroll | q to quit</span>
	</div>
}
