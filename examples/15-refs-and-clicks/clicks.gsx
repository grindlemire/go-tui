package main

import (
	"fmt"
	tui "github.com/grindlemire/go-tui"
)

type colorMixer struct {
	red   *tui.State[int]
	green *tui.State[int]
	blue  *tui.State[int]

	redUpBtn   *tui.Ref
	redDnBtn   *tui.Ref
	greenUpBtn *tui.Ref
	greenDnBtn *tui.Ref
	blueUpBtn  *tui.Ref
	blueDnBtn  *tui.Ref
}

func ColorMixer() *colorMixer {
	return &colorMixer{
		red:        tui.NewState(128),
		green:      tui.NewState(64),
		blue:       tui.NewState(200),
		redUpBtn:   tui.NewRef(),
		redDnBtn:   tui.NewRef(),
		greenUpBtn: tui.NewRef(),
		greenDnBtn: tui.NewRef(),
		blueUpBtn:  tui.NewRef(),
		blueDnBtn:  tui.NewRef(),
	}
}

func clamp(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func (c *colorMixer) adjustRed(delta int) {
	c.red.Set(clamp(c.red.Get()+delta, 0, 255))
}

func (c *colorMixer) adjustGreen(delta int) {
	c.green.Set(clamp(c.green.Get()+delta, 0, 255))
}

func (c *colorMixer) adjustBlue(delta int) {
	c.blue.Set(clamp(c.blue.Get()+delta, 0, 255))
}

func (c *colorMixer) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnRune('q', func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnRune('r', func(ke tui.KeyEvent) { c.adjustRed(16) }),
		tui.OnRune('R', func(ke tui.KeyEvent) { c.adjustRed(-16) }),
		tui.OnRune('g', func(ke tui.KeyEvent) { c.adjustGreen(16) }),
		tui.OnRune('G', func(ke tui.KeyEvent) { c.adjustGreen(-16) }),
		tui.OnRune('b', func(ke tui.KeyEvent) { c.adjustBlue(16) }),
		tui.OnRune('B', func(ke tui.KeyEvent) { c.adjustBlue(-16) }),
	}
}

func (c *colorMixer) HandleMouse(me tui.MouseEvent) bool {
	return tui.HandleClicks(me,
		tui.Click(c.redUpBtn, func() { c.adjustRed(16) }),
		tui.Click(c.redDnBtn, func() { c.adjustRed(-16) }),
		tui.Click(c.greenUpBtn, func() { c.adjustGreen(16) }),
		tui.Click(c.greenDnBtn, func() { c.adjustGreen(-16) }),
		tui.Click(c.blueUpBtn, func() { c.adjustBlue(16) }),
		tui.Click(c.blueDnBtn, func() { c.adjustBlue(-16) }),
	)
}

func colorBar(value int) string {
	filled := value * 20 / 255
	bar := ""
	for i := 0; i < 20; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	return bar
}

templ (c *colorMixer) Render() {
	<div class="flex-col p-2 gap-2 border-rounded border-cyan">
		<span class="text-gradient-cyan-magenta font-bold">Color Mixer</span>

		// Color preview
		<div class="flex-col items-center gap-1 border-rounded p-1">
			<span class="text-gradient-cyan-magenta font-bold">Preview</span>
			<div class="bg-gradient-cyan-magenta" height={3} width={30}>
				<span>{" "}</span>
			</div>
			<div class="flex gap-2 justify-center">
				<span class="text-red font-bold">{fmt.Sprintf("R: %d", c.red.Get())}</span>
				<span class="text-green font-bold">{fmt.Sprintf("G: %d", c.green.Get())}</span>
				<span class="text-blue font-bold">{fmt.Sprintf("B: %d", c.blue.Get())}</span>
			</div>
		</div>

		// Color bars
		<div class="flex-col gap-1 border-rounded p-1">
			<div class="flex gap-1">
				<span class="text-red font-bold w-5">Red</span>
				<span class="text-red">{colorBar(c.red.Get())}</span>
				<span class="text-red font-bold">{fmt.Sprintf("%3d", c.red.Get())}</span>
			</div>
			<div class="flex gap-1">
				<span class="text-green font-bold w-5">Grn</span>
				<span class="text-green">{colorBar(c.green.Get())}</span>
				<span class="text-green font-bold">{fmt.Sprintf("%3d", c.green.Get())}</span>
			</div>
			<div class="flex gap-1">
				<span class="text-blue font-bold w-5">Blu</span>
				<span class="text-blue">{colorBar(c.blue.Get())}</span>
				<span class="text-blue font-bold">{fmt.Sprintf("%3d", c.blue.Get())}</span>
			</div>
		</div>

		// Channel controls with refs
		<div class="flex gap-2">
			<div class="flex-col border-rounded p-1 gap-1 items-center" flexGrow={1.0}>
				<span class="font-bold text-red">Red</span>
				<button ref={c.redUpBtn} class="px-2">{" + "}</button>
				<span class="font-bold text-red">{fmt.Sprintf("%d", c.red.Get())}</span>
				<button ref={c.redDnBtn} class="px-2">{" - "}</button>
			</div>
			<div class="flex-col border-rounded p-1 gap-1 items-center" flexGrow={1.0}>
				<span class="font-bold text-green">Green</span>
				<button ref={c.greenUpBtn} class="px-2">{" + "}</button>
				<span class="font-bold text-green">{fmt.Sprintf("%d", c.green.Get())}</span>
				<button ref={c.greenDnBtn} class="px-2">{" - "}</button>
			</div>
			<div class="flex-col border-rounded p-1 gap-1 items-center" flexGrow={1.0}>
				<span class="font-bold text-blue">Blue</span>
				<button ref={c.blueUpBtn} class="px-2">{" + "}</button>
				<span class="font-bold text-blue">{fmt.Sprintf("%d", c.blue.Get())}</span>
				<button ref={c.blueDnBtn} class="px-2">{" - "}</button>
			</div>
		</div>

		<div class="flex justify-center">
			<span class="font-dim">r/g/b increase | R/G/B decrease | click buttons | q quit</span>
		</div>
	</div>
}
