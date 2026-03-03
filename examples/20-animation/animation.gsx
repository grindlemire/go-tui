package main

import (
	"fmt"
	"math"
	"strings"
	"time"

	tui "github.com/grindlemire/go-tui"
)

// Spinner frame sets. Each spinner cycles through its frames on a timer.
var spinnerDots = []string{"⠋", "⠙", "⠚", "⠞", "⠖", "⠦", "⠴", "⠲", "⠳", "⠓"}
var spinnerLine = []string{"┤", "┘", "┴", "└", "├", "┌", "┬", "┐"}
var spinnerCircle = []string{"◜", "◠", "◝", "◞", "◡", "◟"}
var spinnerBraille = []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"}


// Characters for the color wave animation. Each character gets its own color.
var waveChars = []string{"A", "N", "I", "M", "A", "T", "I", "O", "N", "S"}

// easeInOutCubic applies cubic ease-in-out to a linear t in [0, 1].
// Starts slow, accelerates through the middle, and decelerates at the end.
func easeInOutCubic(t float64) float64 {
	if t < 0.5 {
		return 4 * t * t * t
	}
	return 1 - math.Pow(-2*t+2, 3)/2
}

// Fractional block characters for sub-character precision in progress bars.
// Index 0 is empty, 1-7 are partial fills, used for the fractional part.
var barBlocks = []string{" ", "▏", "▎", "▍", "▌", "▋", "▊", "▉"}

// renderBar draws a smooth progress bar using fractional block characters.
// value is 0.0 to 1.0, width is the bar width in characters.
// Uses 8 sub-steps per character for silky fill transitions.
func renderBar(value float64, width int) string {
	if value < 0 {
		value = 0
	}
	if value > 1 {
		value = 1
	}
	totalSteps := float64(width) * 8
	steps := int(value * totalSteps)
	fullBlocks := steps / 8
	remainder := steps % 8

	var b strings.Builder
	b.WriteString(strings.Repeat("█", fullBlocks))
	if fullBlocks < width {
		b.WriteString(barBlocks[remainder])
		b.WriteString(strings.Repeat("░", width-fullBlocks-1))
	}
	return b.String()
}

// hslToRGB converts HSL color values to RGB.
// h is in [0, 360), s and l are in [0, 1].
func hslToRGB(h, s, l float64) (uint8, uint8, uint8) {
	h = math.Mod(h, 360)
	if h < 0 {
		h += 360
	}

	c := (1 - math.Abs(2*l-1)) * s
	x := c * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := l - c/2

	var r, g, b float64
	switch {
	case h < 60:
		r, g, b = c, x, 0
	case h < 120:
		r, g, b = x, c, 0
	case h < 180:
		r, g, b = 0, c, x
	case h < 240:
		r, g, b = 0, x, c
	case h < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}

	return uint8((r + m) * 255), uint8((g + m) * 255), uint8((b + m) * 255)
}

// waveStyle computes the text style for a single character in the color wave.
// Each character gets a different hue based on its index, shifted by the phase.
func waveStyle(charIndex int, phase float64) tui.Style {
	hue := math.Mod(float64(charIndex)*36+phase*60, 360)
	r, g, b := hslToRGB(hue, 1.0, 0.6)
	return tui.NewStyle().Bold().Foreground(tui.RGBColor(r, g, b))
}

// pulseBorderStyle computes a border style that oscillates between cyan and magenta.
// Uses sin() to create a smooth breathing effect.
func pulseBorderStyle(phase float64) tui.Style {
	t := (math.Sin(phase) + 1) / 2
	color := tui.NewGradient(tui.Cyan, tui.Magenta).At(t)
	return tui.NewStyle().Foreground(color)
}

// progressT computes the linear progress value (0 to 1) from elapsed time.
// Cycles: 3 seconds to fill, 1 second hold, then restart.
func progressT(elapsed time.Duration) float64 {
	const cycleDuration = 3.0
	const pauseDuration = 1.0
	total := cycleDuration + pauseDuration
	cycleTime := math.Mod(elapsed.Seconds(), total)
	if cycleTime >= cycleDuration {
		return 1.0
	}
	return cycleTime / cycleDuration
}

type animationApp struct {
	// Spinner: frame index, incremented every 5th tick
	spinnerFrame *tui.State[int]

	// Color wave: phase offset for rainbow shift
	wavePhase *tui.State[float64]

	// Pulsing border: phase for sin() oscillation
	pulsePhase *tui.State[float64]

	// Animation timing
	startTime time.Time
	frame     int
	paused    bool
}

func AnimationApp() *animationApp {
	return &animationApp{
		spinnerFrame: tui.NewState(0),
		wavePhase:    tui.NewState(0.0),
		pulsePhase:   tui.NewState(0.0),
		startTime:    time.Now(),
	}
}

func (a *animationApp) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnRune('q', func(ke tui.KeyEvent) { ke.App().Stop() }),
	}
}

func (a *animationApp) Watchers() []tui.Watcher {
	return []tui.Watcher{
		tui.OnTimer(16*time.Millisecond, a.animate),
	}
}

// animate is the single animation loop driving all four demos.
// Called every ~16ms (60fps). Uses frame counting to run different
// animations at different speeds.
func (a *animationApp) animate() {
	if a.paused {
		return
	}
	a.frame++

	// Spinners: advance every 5 frames (~80ms)
	if a.frame%5 == 0 {
		a.spinnerFrame.Update(func(v int) int { return v + 1 })
	}

	// Wave: advance phase every frame
	a.wavePhase.Update(func(v float64) float64 { return v + 0.05 })

	// Pulse: advance phase every frame (slower than wave)
	a.pulsePhase.Update(func(v float64) float64 { return v + 0.03 })

	// Progress: computed from wall clock in render, no state needed
}

templ (a *animationApp) Render() {
	<div class="flex-col gap-1 p-1">
		<span class="font-bold text-gradient-cyan-magenta">Animation Patterns</span>

		<div class="flex-col border-rounded p-1 gap-1" borderStyle={tui.NewStyle().Foreground(tui.BrightBlack)}>
			<span class="font-bold text-cyan">1. Spinners (Frame Cycling)</span>
			<div class="flex gap-4">
				<div class="flex gap-1">
					<span class="text-cyan">{spinnerDots[a.spinnerFrame.Get()%len(spinnerDots)]}</span>
					<span class="font-dim">Dots</span>
				</div>
				<div class="flex gap-1">
					<span class="text-green">{spinnerLine[a.spinnerFrame.Get()%len(spinnerLine)]}</span>
					<span class="font-dim">Line</span>
				</div>
				<div class="flex gap-1">
					<span class="text-yellow">{spinnerCircle[a.spinnerFrame.Get()%len(spinnerCircle)]}</span>
					<span class="font-dim">Circle</span>
				</div>
				<div class="flex gap-1">
					<span class="text-magenta">{spinnerBraille[a.spinnerFrame.Get()%len(spinnerBraille)]}</span>
					<span class="font-dim">Braille</span>
				</div>
			</div>
		</div>

		<div class="flex-col border-rounded p-1 gap-1" borderStyle={tui.NewStyle().Foreground(tui.BrightBlack)}>
			<span class="font-bold text-cyan">2. Progress Bar (Easing)</span>
			<div class="flex gap-1">
				<span class="font-dim w-8">Linear:</span>
				<span class="text-blue">{renderBar(progressT(time.Since(a.startTime)), 30)}</span>
				<span class="text-blue">{fmt.Sprintf("%3.0f%%", progressT(time.Since(a.startTime))*100)}</span>
			</div>
			<div class="flex gap-1">
				<span class="font-dim w-8">Eased:</span>
				<span class="text-green">{renderBar(easeInOutCubic(progressT(time.Since(a.startTime))), 30)}</span>
				<span class="text-green">{fmt.Sprintf("%3.0f%%", easeInOutCubic(progressT(time.Since(a.startTime)))*100)}</span>
			</div>
		</div>

		<div class="flex-col border-rounded p-1 gap-1" borderStyle={pulseBorderStyle(a.pulsePhase.Get())}>
			<span class="font-bold text-cyan">3. Color Wave + Pulsing Border</span>
			<div class="flex gap-0">
				@for i, ch := range waveChars {
					<span textStyle={waveStyle(i, a.wavePhase.Get())}>{ch}</span>
				}
			</div>
			<span class="font-dim">Border oscillates between cyan and magenta via sin()</span>
		</div>

		<span class="font-dim">q quit</span>
	</div>
}
