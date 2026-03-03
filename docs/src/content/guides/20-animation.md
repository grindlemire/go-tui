# Animation Patterns

## What We're Building

We're going to build a screen that runs four animation techniques side by side: frame-cycling spinners, eased progress bars, a per-character color wave, and a pulsing border. A single `OnTimer(16ms)` callback drives all of them at 60fps.

Concepts used:

- **State** ([Guide 05](state)): reactive `State[T]` for spinner frames, wave phase, and pulse phase
- **Watchers** ([Guide 09](watchers)): `OnTimer` for the animation loop
- **Styling** ([Guide 03](styling)): `tui.RGBColor`, `tui.NewGradient`, `tui.NewStyle`, dynamic `textStyle` and `borderStyle`
- **Layout** ([Guide 04](layout)): nested flex containers, gap, padding

## Project Setup

Create a new directory and initialize the module:

```bash
mkdir animation && cd animation
go mod init animation
go get github.com/grindlemire/go-tui
```

You'll create two files:

- `animation.gsx` -- the component, helpers, and render template
- `main.go` -- the entry point

## The Animation Loop

Terminal animations boil down to updating state on a timer and letting the framework re-render. The simplest approach: one `OnTimer` callback that ticks at your target frame rate, with frame counting to run different animations at different speeds.

Create `animation.gsx` with the struct and timer:

```gsx
package main

import (
    "fmt"
    "math"
    "strings"
    "time"

    tui "github.com/grindlemire/go-tui"
)

type animationApp struct {
    spinnerFrame *tui.State[int]
    wavePhase    *tui.State[float64]
    pulsePhase   *tui.State[float64]
    startTime    time.Time
    frame        int
    paused       bool
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

func (a *animationApp) animate() {
    if a.paused {
        return
    }
    a.frame++

    if a.frame%5 == 0 {
        a.spinnerFrame.Update(func(v int) int { return v + 1 })
    }

    a.wavePhase.Update(func(v float64) float64 { return v + 0.05 })
    a.pulsePhase.Update(func(v float64) float64 { return v + 0.03 })
}
```

The timer fires every 16 milliseconds, giving roughly 60 frames per second. The `animate` callback increments a frame counter and updates three state values:

- `spinnerFrame` advances every 5th frame (~80ms per step), which gives the spinners a readable pace.
- `wavePhase` advances every frame for a fast-moving rainbow shift.
- `pulsePhase` advances at a slower rate for a gentle breathing effect on the border.

The progress bars use `time.Since(a.startTime)` directly in the render method instead of state, so they run on wall clock time with no accumulation drift.

And `main.go`:

```go
package main

import (
    "fmt"
    "os"

    tui "github.com/grindlemire/go-tui"
)

func main() {
    app, err := tui.NewApp(
        tui.WithRootComponent(AnimationApp()),
    )
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    defer app.Close()

    if err := app.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

## Pattern 1: Spinners

The simplest animation pattern is cycling through an array of Unicode characters. Define several frame sets with different visual styles:

```gsx
var spinnerDots = []string{"⠋", "⠙", "⠚", "⠞", "⠖", "⠦", "⠴", "⠲", "⠳", "⠓"}
var spinnerLine = []string{"┤", "┘", "┴", "└", "├", "┌", "┬", "┐"}
var spinnerCircle = []string{"◜", "◠", "◝", "◞", "◡", "◟"}
var spinnerBraille = []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"}
```

Each array represents one full cycle. The render template indexes into them with modular arithmetic:

```gsx
<span class="text-cyan">{spinnerDots[a.spinnerFrame.Get()%len(spinnerDots)]}</span>
```

The `%len(...)` wraps around at the end of the array, so the spinner loops forever. Different array lengths produce different cycle speeds from the same frame counter.

## Pattern 2: Eased Progress Bars

A linear progress bar moves at a constant rate, which looks stiff. An easing function remaps the progress value so the bar starts slow, speeds up through the middle, and slows down again at the end.

```gsx
func easeInOutCubic(t float64) float64 {
    if t < 0.5 {
        return 4 * t * t * t
    }
    return 1 - math.Pow(-2*t+2, 3)/2
}
```

The input `t` and the output are both in the range [0, 1]. Wrapping the linear progress value with `easeInOutCubic(t)` before passing it to the bar renderer produces visibly smoother motion.

For the bar itself, fractional block characters (`▏▎▍▌▋▊▉█`) give 8 sub-steps per character cell, making the fill transition look smooth instead of jumping one full block at a time:

```gsx
var barBlocks = []string{" ", "▏", "▎", "▍", "▌", "▋", "▊", "▉"}

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
```

The progress timing uses wall clock time through a helper that cycles on a 4-second loop (3 seconds to fill, 1 second hold):

```gsx
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
```

The render template shows both bars side by side so you can see the difference:

```gsx
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
```

## Pattern 3: Per-Character Color Wave

Each character in the word "ANIMATIONS" gets its own color based on its position and a phase offset. As the phase advances every frame, the colors shift across the text like a rainbow wave.

The color computation converts HSL to RGB so we can rotate the hue evenly:

```gsx
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

func waveStyle(charIndex int, phase float64) tui.Style {
    hue := math.Mod(float64(charIndex)*36+phase*60, 360)
    r, g, b := hslToRGB(hue, 1.0, 0.6)
    return tui.NewStyle().Bold().Foreground(tui.RGBColor(r, g, b))
}
```

Each character is spaced 36 degrees apart in hue. The `phase*60` term shifts the entire palette over time. In the template, a `@for` loop renders each character as its own `<span>` with a computed `textStyle`:

```gsx
<div class="flex gap-0">
    @for i, ch := range waveChars {
        <span textStyle={waveStyle(i, a.wavePhase.Get())}>{ch}</span>
    }
</div>
```

## Pattern 4: Pulsing Border

The color wave section's border oscillates between cyan and magenta using `math.Sin()` and the gradient API:

```gsx
func pulseBorderStyle(phase float64) tui.Style {
    t := (math.Sin(phase) + 1) / 2
    color := tui.NewGradient(tui.Cyan, tui.Magenta).At(t)
    return tui.NewStyle().Foreground(color)
}
```

`math.Sin()` outputs a value from -1 to 1, which we remap to the 0 to 1 range. `Gradient.At(t)` interpolates between the two colors at that position, so the border fades back and forth between cyan and magenta. In the template, the `borderStyle` attribute accepts this computed style:

```gsx
<div class="flex-col border-rounded p-1 gap-1" borderStyle={pulseBorderStyle(a.pulsePhase.Get())}>
```

## Full Code

Here's the complete `animation.gsx`:

```gsx
package main

import (
    "fmt"
    "math"
    "strings"
    "time"

    tui "github.com/grindlemire/go-tui"
)

var spinnerDots = []string{"⠋", "⠙", "⠚", "⠞", "⠖", "⠦", "⠴", "⠲", "⠳", "⠓"}
var spinnerLine = []string{"┤", "┘", "┴", "└", "├", "┌", "┬", "┐"}
var spinnerCircle = []string{"◜", "◠", "◝", "◞", "◡", "◟"}
var spinnerBraille = []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"}

var waveChars = []string{"A", "N", "I", "M", "A", "T", "I", "O", "N", "S"}

func easeInOutCubic(t float64) float64 {
    if t < 0.5 {
        return 4 * t * t * t
    }
    return 1 - math.Pow(-2*t+2, 3)/2
}

var barBlocks = []string{" ", "▏", "▎", "▍", "▌", "▋", "▊", "▉"}

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

func waveStyle(charIndex int, phase float64) tui.Style {
    hue := math.Mod(float64(charIndex)*36+phase*60, 360)
    r, g, b := hslToRGB(hue, 1.0, 0.6)
    return tui.NewStyle().Bold().Foreground(tui.RGBColor(r, g, b))
}

func pulseBorderStyle(phase float64) tui.Style {
    t := (math.Sin(phase) + 1) / 2
    color := tui.NewGradient(tui.Cyan, tui.Magenta).At(t)
    return tui.NewStyle().Foreground(color)
}

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
    spinnerFrame *tui.State[int]
    wavePhase    *tui.State[float64]
    pulsePhase   *tui.State[float64]
    startTime    time.Time
    frame        int
    paused       bool
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

func (a *animationApp) animate() {
    if a.paused {
        return
    }
    a.frame++

    if a.frame%5 == 0 {
        a.spinnerFrame.Update(func(v int) int { return v + 1 })
    }

    a.wavePhase.Update(func(v float64) float64 { return v + 0.05 })
    a.pulsePhase.Update(func(v float64) float64 { return v + 0.03 })
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
```

And the complete `main.go`:

```go
package main

import (
    "fmt"
    "os"

    tui "github.com/grindlemire/go-tui"
)

func main() {
    app, err := tui.NewApp(
        tui.WithRootComponent(AnimationApp()),
    )
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    defer app.Close()

    if err := app.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

Generate and run:

```bash
tui generate ./...
go run .
```

The finished animation demo with all three sections running:

![Animation Patterns screenshot](/guides/20.png)

## Next Steps

A few ways you could extend this:

- Add a pause/resume toggle with a key binding ([Events Guide](events))
- Use `tui.Batch()` to group multiple state updates into a single render pass ([State Guide](state))
- Combine animations with scrollable content for a loading indicator above a log feed ([Scrolling Guide](scrolling))
