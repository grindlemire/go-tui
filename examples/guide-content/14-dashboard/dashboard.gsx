package main

import (
	"fmt"
	"math/rand"
	"time"

	tui "github.com/grindlemire/go-tui"
)

type dashboardApp struct {
	cpu       *tui.State[int]
	mem       *tui.State[int]
	disk      *tui.State[int]
	netIn     *tui.State[int]
	netOut    *tui.State[int]
	sparkIn   *tui.State[[]int]
	sparkOut  *tui.State[[]int]
	events    *tui.State[[]string]
	eventCh   <-chan string
	scrollY   *tui.State[int]
	eventsRef *tui.Ref
}

func Dashboard(eventCh <-chan string) *dashboardApp {
	return &dashboardApp{
		cpu:       tui.NewState(45),
		mem:       tui.NewState(62),
		disk:      tui.NewState(38),
		netIn:     tui.NewState(142),
		netOut:    tui.NewState(89),
		sparkIn:   tui.NewState([]int{3, 5, 4, 6, 7, 5, 4, 3, 5, 6, 7, 8, 6, 5, 4, 3, 5, 6, 7, 5}),
		sparkOut:  tui.NewState([]int{2, 3, 4, 3, 5, 4, 3, 2, 3, 4, 5, 6, 4, 3, 2, 3, 4, 5, 4, 3}),
		events:    tui.NewState([]string{}),
		eventCh:   eventCh,
		scrollY:   tui.NewState(0),
		eventsRef: tui.NewRef(),
	}
}

func (d *dashboardApp) scrollBy(delta int) {
	el := d.eventsRef.El()
	if el == nil {
		return
	}
	_, maxY := el.MaxScroll()
	newY := d.scrollY.Get() + delta
	if newY < 0 {
		newY = 0
	} else if newY > maxY {
		newY = maxY
	}
	d.scrollY.Set(newY)
}

func (d *dashboardApp) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnRune('q', func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnRune('j', func(ke tui.KeyEvent) { d.scrollBy(1) }),
		tui.OnRune('k', func(ke tui.KeyEvent) { d.scrollBy(-1) }),
		tui.OnKey(tui.KeyDown, func(ke tui.KeyEvent) { d.scrollBy(1) }),
		tui.OnKey(tui.KeyUp, func(ke tui.KeyEvent) { d.scrollBy(-1) }),
	}
}

func (d *dashboardApp) HandleMouse(me tui.MouseEvent) bool {
	switch me.Button {
	case tui.MouseWheelUp:
		d.scrollBy(-1)
		return true
	case tui.MouseWheelDown:
		d.scrollBy(1)
		return true
	}
	return false
}

func (d *dashboardApp) Watchers() []tui.Watcher {
	return []tui.Watcher{
		tui.OnTimer(500*time.Millisecond, d.updateMetrics),
		tui.Watch(d.eventCh, d.addEvent),
	}
}

func (d *dashboardApp) updateMetrics() {
	d.cpu.Set(clampVal(d.cpu.Get()+rand.Intn(11)-5, 5, 95))
	d.mem.Set(clampVal(d.mem.Get()+rand.Intn(7)-3, 20, 90))
	d.disk.Set(clampVal(d.disk.Get()+rand.Intn(3)-1, 20, 80))
	d.netIn.Set(clampVal(d.netIn.Get()+rand.Intn(41)-20, 50, 300))
	d.netOut.Set(clampVal(d.netOut.Get()+rand.Intn(31)-15, 30, 200))

	inData := d.sparkIn.Get()
	inData = append(inData[1:], d.netIn.Get()/30)
	d.sparkIn.Set(inData)

	outData := d.sparkOut.Get()
	outData = append(outData[1:], d.netOut.Get()/30)
	d.sparkOut.Set(outData)
}

func (d *dashboardApp) addEvent(event string) {
	current := d.events.Get()
	ts := time.Now().Format("15:04:05")
	entry := fmt.Sprintf("%s  %s", ts, event)
	current = append(current, entry)
	if len(current) > 50 {
		current = current[len(current)-50:]
	}
	d.events.Set(current)

	// Auto-scroll to bottom
	el := d.eventsRef.El()
	if el != nil {
		_, maxY := el.MaxScroll()
		d.scrollY.Set(maxY + 1)
	}
}

func clampVal(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func metricBar(value, max int) string {
	width := 20
	filled := value * width / max
	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	return bar
}

func metricColor(value int) string {
	if value >= 80 {
		return "text-red font-bold"
	}
	if value >= 60 {
		return "text-yellow"
	}
	return "text-green"
}

func sparkline(data []int) string {
	blocks := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
	maxVal := 1
	for _, v := range data {
		if v > maxVal {
			maxVal = v
		}
	}
	s := ""
	for _, v := range data {
		idx := v * 7 / maxVal
		if idx > 7 {
			idx = 7
		}
		s += string(blocks[idx])
	}
	return s
}

func produceEvents(ch chan<- string, stopCh <-chan struct{}) {
	defer close(ch)
	events := []string{
		"Deploy completed",
		"Health check passed",
		"New connection from 10.0.0.5",
		"Cache invalidated",
		"Backup complete",
		"Certificate renewed",
		"Config reloaded",
		"Scale up: 3 replicas",
		"Alert cleared: cpu",
		"Metrics exported",
	}
	for {
		delay := time.Duration(2000+rand.Intn(3000)) * time.Millisecond
		select {
		case <-stopCh:
			return
		case <-time.After(delay):
		}
		event := events[rand.Intn(len(events))]
		select {
		case <-stopCh:
			return
		case ch <- event:
		}
	}
}

templ (d *dashboardApp) Render() {
	<div class="flex-col p-1 gap-1 h-full border-rounded border-cyan">
		<div class="flex justify-center shrink-0">
			<span class="text-gradient-cyan-magenta font-bold">Dashboard</span>
		</div>

		<div class="flex gap-1 shrink-0">
			<div class="flex-col border-rounded p-1 gap-1" flexGrow={1.0}>
				<span class="text-gradient-cyan-magenta font-bold">CPU</span>
				<span class={metricColor(d.cpu.Get())}>{metricBar(d.cpu.Get(), 100)}</span>
				<span class={metricColor(d.cpu.Get()) + " font-bold"}>{fmt.Sprintf("%d%%", d.cpu.Get())}</span>
			</div>
			<div class="flex-col border-rounded p-1 gap-1" flexGrow={1.0}>
				<span class="text-gradient-cyan-magenta font-bold">Memory</span>
				<span class={metricColor(d.mem.Get())}>{metricBar(d.mem.Get(), 100)}</span>
				<span class={metricColor(d.mem.Get()) + " font-bold"}>{fmt.Sprintf("%d%%", d.mem.Get())}</span>
			</div>
			<div class="flex-col border-rounded p-1 gap-1" flexGrow={1.0}>
				<span class="text-gradient-cyan-magenta font-bold">Disk</span>
				<span class={metricColor(d.disk.Get())}>{metricBar(d.disk.Get(), 100)}</span>
				<span class={metricColor(d.disk.Get()) + " font-bold"}>{fmt.Sprintf("%d%%", d.disk.Get())}</span>
			</div>
		</div>

		<div class="flex gap-1 flex-grow">
			<div class="flex-col border-rounded p-1 gap-1" flexGrow={1.0}>
				<span class="text-gradient-cyan-magenta font-bold">Network Traffic</span>
				<div class="flex gap-1">
					<span class="font-dim">In: </span>
					<span class="text-cyan">{sparkline(d.sparkIn.Get())}</span>
				</div>
				<div class="flex gap-1">
					<span class="font-dim">Out:</span>
					<span class="text-magenta">{sparkline(d.sparkOut.Get())}</span>
				</div>
				<div class="flex gap-2">
					<span class="text-cyan font-bold">{fmt.Sprintf("In: %d MB/s", d.netIn.Get())}</span>
					<span class="text-magenta font-bold">{fmt.Sprintf("Out: %d MB/s", d.netOut.Get())}</span>
				</div>
			</div>

			<div
				ref={d.eventsRef}
				class="flex-col border-rounded p-1 gap-1"
				flexGrow={1.0}
				scrollable={tui.ScrollVertical}
				scrollOffset={0, d.scrollY.Get()}
			>
				<span class="text-gradient-cyan-magenta font-bold">Recent Events</span>
				@for _, event := range d.events.Get() {
					<span class="text-green">{event}</span>
				}
				@if len(d.events.Get()) == 0 {
					<span class="font-dim">Waiting for events...</span>
				}
			</div>
		</div>

		<div class="flex justify-center shrink-0">
			<span class="font-dim">j/k scroll events | q to quit</span>
		</div>
	</div>
}
