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
		tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { tui.Stop() }),
		tui.OnRune('q', func(ke tui.KeyEvent) { tui.Stop() }),
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
		class="flex-col gap-1 p-2 border-rounded h-full"
		scrollable={tui.ScrollVertical}
		scrollOffset={0, s.scrollY.Get()}
	>
		// Text Styles
		<div class="flex-col border-white border-single p-0">
			<span class="font-bold">Text Styles</span>
			<div class="flex-row gap-1">
				<span class="font-bold">Bold text</span>
				<span class="font-dim">Dim text</span>
				<span class="italic">Italic text</span>
				<span class="underline">Underlined text</span>
				<span class="strikethrough">Strikethrough text</span>
				<span class="reverse">Reverse text</span>
				<span class="font-bold italic underline">Bold+Italic+Underline</span>
			</div>
		</div>
		<hr />
		// Text Colors
		<div class="flex-col border-white border-single p-0">
			<span class="font-bold">Text Colors</span>
			<div class="flex-row gap-1">
				<span class="text-red">Red</span>
				<span class="text-green">Green</span>
				<span class="text-blue">Blue</span>
				<span class="text-cyan">Cyan</span>
				<span class="text-magenta">Magenta</span>
				<span class="text-yellow">Yellow</span>
				<span class="text-white">White</span>
			</div>
		</div>
		<hr />
		// Bright Text Colors
		<div class="flex-col border-white border-single p-0">
			<span class="font-bold">Bright Text Colors</span>
			<div class="flex-row gap-1">
				<span class="text-bright-red">Red</span>
				<span class="text-bright-green">Green</span>
				<span class="text-bright-blue">Blue</span>
				<span class="text-bright-cyan">Cyan</span>
				<span class="text-bright-magenta">Magenta</span>
				<span class="text-bright-yellow">Yellow</span>
				<span class="text-bright-white">White</span>
			</div>
		</div>
		<hr />
		// Background Colors
		<div class="flex-col border-white border-single p-0">
			<span class="font-bold">Background Colors</span>
			<div class="flex-row gap-1">
				<span class="bg-red">Red</span>
				<span class="bg-green">Green</span>
				<span class="bg-blue">Blue</span>
				<span class="bg-cyan">Cyan</span>
				<span class="bg-magenta">Magenta</span>
				<span class="bg-yellow">Yellow</span>
				<span class="bg-white">White</span>
			</div>
		</div>
		<hr />
		// Bright Background Colors
		<div class="flex-col border-white border-single p-0">
			<span class="font-bold">Bright Background Colors</span>
			<div class="flex-row gap-1">
				<span class="bg-bright-red">Red</span>
				<span class="bg-bright-green">Green</span>
				<span class="bg-bright-blue">Blue</span>
				<span class="bg-bright-cyan">Cyan</span>
				<span class="bg-bright-magenta">Magenta</span>
				<span class="bg-bright-yellow">Yellow</span>
				<span class="bg-bright-white">White</span>
			</div>
		</div>
		<hr />
		// Combined Foreground+Background
		<div class="flex-col border-white border-single p-0">
			<span class="font-bold">Combined Foreground+Background</span>
			<div class="flex-row gap-1">
				<span class="text-white bg-red">Error</span>
				<span class="text-black bg-yellow">Warning</span>
				<span class="text-white bg-green">Success</span>
				<span class="text-white bg-blue">Info</span>
				<span class="font-bold text-black bg-cyan">Highlight</span>
			</div>
		</div>
		<hr />
		// Border Styles
		<div class="flex-col border-white border-single p-0">
			<span class="font-bold">Border Styles</span>
			<div class="flex-row gap-1">
				<div class="border-single">
					<span>Single</span>
				</div>
				<div class="border-double">
					<span>Double</span>
				</div>
				<div class="border-rounded">
					<span>Rounded</span>
				</div>
				<div class="border-thick">
					<span>Thick</span>
				</div>
			</div>
		</div>
		<hr />
		// Colored Borders
		<div class="flex-col border-white border-single p-0">
			<span class="font-bold">Colored Borders</span>
			<div class="flex-row gap-1">
				<div class="border-rounded border-red">
					<span>Red</span>
				</div>
				<div class="border-rounded border-green">
					<span>Green</span>
				</div>
				<div class="border-rounded border-blue">
					<span>Blue</span>
				</div>
				<div class="border-rounded border-cyan">
					<span>Cyan</span>
				</div>
				<div class="border-rounded border-magenta">
					<span>Magenta</span>
				</div>
				<div class="border-rounded border-yellow">
					<span>Yellow</span>
				</div>
				<div class="border-rounded border-white">
					<span>White</span>
				</div>
				<div class="border-rounded border-black">
					<span>Black</span>
				</div>
			</div>
		</div>
		<hr />
		// Text Gradients
		<div class="flex-col border-white border-single p-0">
			<span class="font-bold">Text Gradients</span>
			<div class="flex-col gap-1">
				<div class="flex-row gap-1">
					<span class="text-gradient-red-blue">Red to Blue</span>
					<span class="text-gradient-cyan-magenta">Cyan to Magenta</span>
					<span class="text-gradient-yellow-red">Yellow to Red</span>
					<span class="text-gradient-green-blue">Green to Blue</span>
				</div>
				<div class="flex-row gap-1">
					<span class="text-gradient-red-blue-v">Vertical</span>
					<span class="text-gradient-cyan-magenta-dd">Diagonal Down</span>
					<span class="text-gradient-yellow-red-du">Diagonal Up</span>
				</div>
				<div class="flex-row gap-1">
					<span class="text-gradient-bright-red-bright-blue">Bright Red to Blue</span>
					<span class="text-gradient-bright-cyan-bright-magenta">Bright Cyan to Magenta</span>
				</div>
			</div>
		</div>
		<hr />
		// Background Gradients
		<div class="flex-col border-white border-single p-0">
			<span class="font-bold">Background Gradients</span>
			<div class="flex-col gap-1">
				<div class="flex-row gap-1">
					<div class="bg-gradient-red-blue p-1">
						<span>Horizontal</span>
					</div>
					<div class="bg-gradient-cyan-magenta-v p-1">
						<span>Vertical</span>
					</div>
					<div class="bg-gradient-yellow-red-dd p-1">
						<span>Diagonal Down</span>
					</div>
					<div class="bg-gradient-green-blue-du p-1">
						<span>Diagonal Up</span>
					</div>
				</div>
				<div class="flex-row gap-1">
					<div class="bg-gradient-bright-red-bright-blue p-1">
						<span class="text-white">Bright Colors</span>
					</div>
					<div class="bg-gradient-white-black p-1">
						<span>White to Black</span>
					</div>
					<div class="bg-gradient-black-white p-1">
						<span class="text-white">Black to White</span>
					</div>
				</div>
			</div>
		</div>
		<hr />
		// Border Gradients
		<div class="flex-col border-white border-single p-0">
			<span class="font-bold">Border Gradients</span>
			<div class="flex-row gap-1">
				<div class="border-rounded border-gradient-red-blue p-1">
					<span>Red to Blue</span>
				</div>
				<div class="border-single border-gradient-cyan-magenta p-1">
					<span>Cyan to Magenta</span>
				</div>
				<div class="border-double border-gradient-yellow-red p-1">
					<span>Yellow to Red</span>
				</div>
				<div class="border-thick border-gradient-green-blue p-1">
					<span>Green to Blue</span>
				</div>
			</div>
		</div>
		<hr />
		// Combined Gradients
		<div class="flex-col border-white border-single p-0">
			<span class="font-bold">Combined Gradients</span>
			<div class="flex-col gap-1">
				<div class="bg-gradient-red-blue border-gradient-yellow-red border-rounded p-1">
					<span class="text-gradient-white-black">Text+Bg+Border</span>
				</div>
				<div class="bg-gradient-cyan-magenta-v border-gradient-green-blue border-single p-1">
					<span class="text-gradient-bright-red-bright-blue">All Gradients</span>
				</div>
			</div>
		</div>
		<hr />
		<span class="font-dim">Press q to quit</span>
	</div>
}
