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
		lastEvent:  tui.NewState("(waiting)"),
		eventCount: tui.NewState(0),
	}
	events.Subscribe(func(desc string) {
		e.lastEvent.Set(desc)
		e.eventCount.Set(e.eventCount.Get() + 1)
	})
	return e
}

templ (e *eventInspector) Render() {
	<div
		class="border-single p-1 flex-col gap-1"
		flexGrow={1.0}>
		<span class="text-gradient-yellow-green font-bold">{"Event Inspector"}</span>
		<div class="flex gap-1 items-center">
			<span class="font-dim">Last:</span>
			<span class="text-yellow font-bold">{e.lastEvent.Get()}</span>
		</div>
		<div class="flex gap-1 items-center">
			<span class="font-dim">Total:</span>
			<span class="text-green font-bold">{fmt.Sprintf("%d", e.eventCount.Get())}</span>
		</div>
		<span class="font-dim">{"Events from all components"}</span>
	</div>
}
