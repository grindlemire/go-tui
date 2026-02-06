package main

import (
	"fmt"
	tui "github.com/grindlemire/go-tui"
)

type eventInspector struct {
	lastEvent  *tui.State[string]
	eventCount *tui.State[int]
}

func EventInspector(events *Events[string]) *eventInspector {
	e := &eventInspector{
		lastEvent:  tui.NewState("(none)"),
		eventCount: tui.NewState(0),
	}

	events.Subscribe(func(event string) {
		e.lastEvent.Set(event)
		e.eventCount.Set(e.eventCount.Get() + 1)
	})

	return e
}

templ (e *eventInspector) Render() {
	<div class="border-single p-1 flex-col gap-1" flexGrow={1.0}>
		<span class="text-gradient-magenta-cyan font-bold">{"Event Inspector"}</span>
		<div class="flex gap-1 items-center">
			<span class="font-dim">Last:</span>
			<span class="text-magenta font-bold">{e.lastEvent.Get()}</span>
		</div>
		<div class="flex gap-1 items-center">
			<span class="font-dim">Count:</span>
			<span class="text-cyan font-bold">{fmt.Sprintf("%d", e.eventCount.Get())}</span>
		</div>
		<span>{" "}</span>
		<span class="font-dim">{"events from all components"}</span>
	</div>
}
