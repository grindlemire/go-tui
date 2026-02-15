package main

import (
	"fmt"
	"time"
	tui "github.com/grindlemire/go-tui"
)

type dashboardApp struct {
	cardWidth  *tui.State[int]
	cardHeight *tui.State[int]
	growing    bool
}

var (
	_ tui.WatcherProvider = (*dashboardApp)(nil)
)

func Dashboard() *dashboardApp {
	return &dashboardApp{
		cardWidth:  tui.NewState(25),
		cardHeight: tui.NewState(6),
		growing:    true,
	}
}

func (d *dashboardApp) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnRune('q', func(ke tui.KeyEvent) { ke.App().Stop() }),
	}
}

func (d *dashboardApp) Watchers() []tui.Watcher {
	return []tui.Watcher{
		tui.OnTimer(50*time.Millisecond, d.animate),
	}
}

func (d *dashboardApp) animate() {
	minW, maxW := 25, 50
	minH, maxH := 6, 12

	w := d.cardWidth.Get()
	if d.growing {
		w++
		if w >= maxW {
			d.growing = false
		}
	} else {
		w--
		if w <= minW {
			d.growing = true
		}
	}
	d.cardWidth.Set(w)
	d.cardHeight.Set(minH + (w-minW)*(maxH-minH)/(maxW-minW))
}

func cardSize(w, h int) string {
	return fmt.Sprintf("%dx%d", w, h)
}

templ (d *dashboardApp) Render() {
	<div class="flex-col h-full">
		<div class="flex justify-center items-center border-single text-blue" height={3} flexShrink={0}>
			<span class="font-bold text-white">Dashboard</span>
		</div>
		<div class="flex grow">
			<div class="flex-col p-1 gap-1 border-single text-magenta" width={20} flexShrink={0}>
				<span class="font-bold text-magenta">Menu</span>
				<span class="text-green">{"> Overview"}</span>
				<span class="text-white">{"  Settings"}</span>
				<span class="text-white">{"  Help"}</span>
			</div>
			<div class="flex-col grow justify-center items-center">
				<div class="border-rounded p-1 flex-col gap-1 justify-center items-center text-cyan" width={d.cardWidth.Get()} height={d.cardHeight.Get()}>
					<span class="font-bold text-cyan">Status Card</span>
					<span class="text-green">Systems Online</span>
					<span class="font-dim">{cardSize(d.cardWidth.Get(), d.cardHeight.Get())}</span>
				</div>
			</div>
		</div>
		<div class="flex justify-center p-1 shrink-0">
			<span class="font-dim">Press q to quit</span>
		</div>
	</div>
}
