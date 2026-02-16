package main

import tui "github.com/grindlemire/go-tui"

type stylingApp struct {
	scrollY *tui.State[int]
	content *tui.Ref
}

func Styling() *stylingApp {
	return &stylingApp{
		scrollY: tui.NewState(0),
		content: tui.NewRef(),
	}
}

func (s *stylingApp) scrollBy(delta int) {
	el := s.content.El()
	if el == nil {
		return
	}
	_, maxY := el.MaxScroll()
	newY := s.scrollY.Get() + delta
	if newY < 0 {
		newY = 0
	} else if newY > maxY {
		newY = maxY
	}
	s.scrollY.Set(newY)
}

func (s *stylingApp) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnRune('q', func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnRune('j', func(ke tui.KeyEvent) { s.scrollBy(1) }),
		tui.OnRune('k', func(ke tui.KeyEvent) { s.scrollBy(-1) }),
		tui.OnKey(tui.KeyDown, func(ke tui.KeyEvent) { s.scrollBy(1) }),
		tui.OnKey(tui.KeyUp, func(ke tui.KeyEvent) { s.scrollBy(-1) }),
	}
}

func (s *stylingApp) HandleMouse(me tui.MouseEvent) bool {
	switch me.Button {
	case tui.MouseWheelUp:
		s.scrollBy(-1)
		return true
	case tui.MouseWheelDown:
		s.scrollBy(1)
		return true
	}
	return false
}

templ (s *stylingApp) Render() {
	<div
		ref={s.content}
		class="flex-col gap-1 p-2 h-full"
		scrollable={tui.ScrollVertical}
		scrollOffset={0, s.scrollY.Get()}
	>
		<span class="text-gradient-cyan-magenta font-bold">Style Guide</span>

		// Row 1: Text Colors + Text Styles
		<div class="flex gap-1">
			<div class="flex-col border-rounded p-1 gap-1">
				<span class="text-gradient-cyan-magenta font-bold">Text Colors</span>
				<div class="flex gap-1">
					<span class="text-red">red</span>
					<span class="text-green">green</span>
					<span class="text-blue">blue</span>
					<span class="text-cyan">cyan</span>
					<span class="text-magenta">magenta</span>
					<span class="text-yellow">yellow</span>
					<span class="text-white">white</span>
				</div>
				<div class="flex gap-1">
					<span class="text-bright-red">bright-red</span>
					<span class="text-bright-green">bright-green</span>
					<span class="text-bright-blue">bright-blue</span>
					<span class="text-bright-cyan">bright-cyan</span>
					<span class="text-bright-magenta">bright-magenta</span>
				</div>
			</div>
			<div class="flex-col border-rounded p-1 gap-1">
				<span class="text-gradient-cyan-magenta font-bold">Text Styles</span>
				<div class="flex gap-1">
					<span class="font-bold">Bold</span>
					<span class="font-dim">Dim</span>
					<span class="italic">Italic</span>
				</div>
				<div class="flex gap-1">
					<span class="underline">Underline</span>
					<span class="strikethrough">Strikethrough</span>
					<span class="reverse">Reverse</span>
				</div>
				<span class="font-bold italic underline">Combined styles</span>
			</div>
		</div>

		// Row 2: Borders + Gradients
		<div class="flex gap-1">
			<div class="flex-col border-rounded p-1 gap-1">
				<span class="text-gradient-cyan-magenta font-bold">Borders</span>
				<div class="flex gap-1">
					<div class="border-single p-1">
						<span>Single</span>
					</div>
					<div class="border-double p-1">
						<span>Double</span>
					</div>
					<div class="border-rounded p-1">
						<span>Rounded</span>
					</div>
					<div class="border-thick p-1">
						<span>Thick</span>
					</div>
				</div>
				<div class="flex gap-1">
					<div class="border-rounded border-red p-1">
						<span>red</span>
					</div>
					<div class="border-rounded border-green p-1">
						<span>green</span>
					</div>
					<div class="border-rounded border-blue p-1">
						<span>blue</span>
					</div>
					<div class="border-rounded border-cyan p-1">
						<span>cyan</span>
					</div>
					<div class="border-rounded border-magenta p-1">
						<span>magenta</span>
					</div>
					<div class="border-rounded border-yellow p-1">
						<span>yellow</span>
					</div>
				</div>
			</div>
			<div class="flex-col border-rounded p-1 gap-1">
				<span class="text-gradient-cyan-magenta font-bold">Gradients</span>
				<span class="text-gradient-red-blue">Horizontal red to blue</span>
				<span class="text-gradient-cyan-magenta">Horizontal cyan to magenta</span>
				<span class="text-gradient-yellow-red-v">Vertical yellow to red</span>
				<span class="text-gradient-green-blue-dd">Diagonal green to blue</span>
			</div>
		</div>

		// Row 3: Backgrounds
		<div class="flex-col border-rounded p-1 gap-1">
			<span class="text-gradient-cyan-magenta font-bold">Backgrounds</span>
			<div class="flex gap-1">
				<span class="bg-red text-white px-1">red</span>
				<span class="bg-green text-white px-1">green</span>
				<span class="bg-blue text-white px-1">blue</span>
				<span class="bg-cyan text-black px-1">cyan</span>
				<span class="bg-magenta text-white px-1">magenta</span>
				<span class="bg-yellow text-black px-1">yellow</span>
			</div>
			<div class="flex gap-1">
				<span class="text-white bg-red font-bold px-1">Error</span>
				<span class="text-black bg-yellow font-bold px-1">Warning</span>
				<span class="text-white bg-green font-bold px-1">Success</span>
				<span class="text-white bg-blue font-bold px-1">Info</span>
				<span class="font-bold text-black bg-cyan px-1">Highlight</span>
			</div>
		</div>

		// Row 4: Background + Border Gradients
		<div class="flex gap-1">
			<div class="flex-col border-rounded p-1 gap-1">
				<span class="text-gradient-cyan-magenta font-bold">Background Gradients</span>
				<div class="flex gap-1">
					<div class="bg-gradient-red-blue p-1">
						<span class="text-white">Horizontal</span>
					</div>
					<div class="bg-gradient-cyan-magenta-v p-1">
						<span class="text-white">Vertical</span>
					</div>
					<div class="bg-gradient-yellow-red-dd p-1">
						<span>Diagonal</span>
					</div>
				</div>
			</div>
			<div class="flex-col border-rounded p-1 gap-1">
				<span class="text-gradient-cyan-magenta font-bold">Border Gradients</span>
				<div class="flex gap-1">
					<div class="border-rounded border-gradient-red-blue p-1">
						<span>Red-Blue</span>
					</div>
					<div class="border-single border-gradient-cyan-magenta p-1">
						<span>Cyan-Magenta</span>
					</div>
					<div class="border-double border-gradient-yellow-red p-1">
						<span>Yellow-Red</span>
					</div>
				</div>
			</div>
		</div>

		// Row 5: Combined
		<div class="flex-col border-rounded p-1 gap-1">
			<span class="text-gradient-cyan-magenta font-bold">Combined Gradients</span>
			<div class="flex gap-1">
				<div class="bg-gradient-red-blue border-gradient-yellow-red border-rounded p-1">
					<span class="text-gradient-white-black">Text + Bg + Border</span>
				</div>
				<div class="bg-gradient-cyan-magenta-v border-gradient-green-blue border-single p-1">
					<span class="text-gradient-bright-red-bright-blue">All Gradients</span>
				</div>
			</div>
		</div>

		<span class="font-dim">j/k to scroll | q to quit</span>
	</div>
}
