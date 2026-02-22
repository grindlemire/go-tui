package main

import (
	"fmt"

	tui "github.com/grindlemire/go-tui"
)

// Pure components

templ Badge(label string, color string) {
	<span class={color + " font-bold px-1"}>{label}</span>
}

templ StatusLine(label string, value string) {
	<div class="flex gap-1">
		<span class="font-dim">{label}</span>
		<span class="text-cyan font-bold">{value}</span>
	</div>
}

templ Card(title string) {
	<div class="border-rounded p-1 flex-col gap-1 w-full" flexGrow={1.0}>
		<span class="text-gradient-cyan-magenta font-bold">{title}</span>
		<hr class="border-single" />
		{children...}
	</div>
}

// Tab content components

templ OverviewTab() {
	<div class="flex gap-1">
		@Card("System") {
			@StatusLine("CPU:", "42%")
			@StatusLine("Memory:", "1.2 GB")
			@StatusLine("Disk:", "68%")
		}
		@Card("Network") {
			@StatusLine("In:", "12 MB/s")
			@StatusLine("Out:", "3 MB/s")
			@StatusLine("Latency:", "24ms")
		}
	</div>
}

templ MetricsTab() {
	<div class="flex gap-1">
		@Card("Performance") {
			@StatusLine("Requests:", "1.2k/s")
			@StatusLine("P99:", "145ms")
			@StatusLine("Errors:", "0.02%")
		}
		@Card("Storage") {
			@StatusLine("Used:", "42 GB")
			@StatusLine("Free:", "118 GB")
			@StatusLine("IOPS:", "3.4k")
		}
	</div>
}

templ LogsTab() {
	<div class="flex gap-1">
		@Card("Application") {
			@StatusLine("Level:", "INFO")
			@StatusLine("Rate:", "84/min")
			@StatusLine("Errors:", "2")
		}
		@Card("Security") {
			@StatusLine("Auth:", "OK")
			@StatusLine("Blocked:", "17")
			@StatusLine("Alerts:", "0")
		}
	</div>
}

// Struct component

type dashboard struct {
	selected *tui.State[int]
	tabs     []string
}

func Dashboard() *dashboard {
	return &dashboard{
		selected: tui.NewState(0),
		tabs:     []string{"Overview", "Metrics", "Logs"},
	}
}

func (d *dashboard) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnKey(tui.KeyTab, func(ke tui.KeyEvent) {
			d.selected.Update(func(v int) int {
				return (v + 1) % len(d.tabs)
			})
		}),
	}
}

templ (d *dashboard) Render() {
	<div class="flex-col p-1 gap-1 border-rounded border-cyan">
		<div class="flex gap-2">
			@for i, tab := range d.tabs {
				@if i == d.selected.Get() {
					@Badge(tab, "text-cyan")
				} @else {
					<span class="font-dim">{tab}</span>
				}
			}
		</div>

		@if d.selected.Get() == 0 {
			@OverviewTab()
		} @else @if d.selected.Get() == 1 {
			@MetricsTab()
		} @else {
			@LogsTab()
		}

		<div class="flex justify-center">
			<span class="font-dim">{fmt.Sprintf("Tab: switch tabs | esc: quit | Viewing: %s", d.tabs[d.selected.Get()])}</span>
		</div>
	</div>
}
