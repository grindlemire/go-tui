package main

import (
    "fmt"
    "time"
    tui "github.com/grindlemire/go-tui"
)

type counterApp struct {
	count   *tui.State[int]
	elapsed *tui.State[int]
	display *tui.Ref
}

func Counter() *counterApp {
	return &counterApp{
		count:   tui.NewState(0),
		elapsed: tui.NewState(0),
		display: tui.NewRef(),
	}
}

func (c *counterApp) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnRune('+', func(ke tui.KeyEvent) {
			c.count.Update(func(v int) int { return v + 1 })
		}),
		tui.OnRune('-', func(ke tui.KeyEvent) {
			c.count.Update(func(v int) int { return v - 1 })
		}),
		tui.OnRune('r', func(ke tui.KeyEvent) { c.count.Set(0) }),
		tui.OnRune('q', func(ke tui.KeyEvent) { ke.App().Stop() }),
	}
}

func (c *counterApp) Watchers() []tui.Watcher {
	return []tui.Watcher{
		tui.OnTimer(time.Second, func() {
			c.elapsed.Update(func(v int) int { return v + 1 })
		}),
	}
}

func formatTime(seconds int) string {
	return fmt.Sprintf("%d:%02d", seconds/60, seconds%60)
}

templ Badge(label string, value string, color string) {
    <div class="flex gap-1">
        <span class="font-dim">{label}</span>
        <span class={"font-bold " + color}>{value}</span>
    </div>
}

templ Card(title string) {
    <div class="flex-col border-rounded p-1 gap-1" flexGrow={1.0}>
        <span class="font-bold text-cyan">{title}</span>
        <hr />
        {children...}
    </div>
}

templ (c *counterApp) Render() {
    <div class="flex-col border-rounded p-1 gap-1">
        <div class="flex justify-between">
            <span class="font-bold text-cyan">Counter</span>
            @Badge("uptime:", formatTime(c.elapsed.Get()), "text-yellow")
        </div>
        <hr />
        <div class="flex gap-2">
            @Card("Count") {
                <span ref={c.display} class="text-cyan font-bold">
                    {fmt.Sprintf("%d", c.count.Get())}
                </span>
            }
            @Card("Status") {
                @if c.count.Get() > 0 {
                    <span class="text-green font-bold">Positive</span>
                } @else @if c.count.Get() < 0 {
                    <span class="text-red font-bold">Negative</span>
                } @else {
                    <span class="text-blue font-bold">Zero</span>
                }
            }
        </div>
        <div class="flex gap-1 justify-center">
            <span class="font-dim">+/-count·r reset·q quit</span>
        </div>
    </div>
}
