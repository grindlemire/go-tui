package main

import (
	"time"
	tui "github.com/grindlemire/go-tui"
)

type interactiveApp struct {
	events    *Events[string]
	registrar *Registrar

	// Child components
	counter        *counter
	timer          *timer
	toggles        *toggles
	eventInspector *eventInspector
}

func Interactive() *interactiveApp {
	events := NewEvents[string]()

	// Create child components
	counterComp := Counter(events)
	timerComp := Timer(events)
	togglesComp := Toggles(events)
	inspectorComp := EventInspector(events)

	// Collect all registrations
	reg := NewRegistrar()

	// App-level bindings
	reg.OnRune('q', func() { tui.Stop() })
	reg.OnKey(tui.KeyEscape, func() { tui.Stop() })
	reg.OnKey(tui.KeyEnter, func() { events.Emit("Enter") })
	reg.OnKey(tui.KeyTab, func() { events.Emit("Tab") })
	reg.OnKey(tui.KeyBackspace, func() { events.Emit("Backspace") })
	reg.OnKey(tui.KeyUp, func() { events.Emit("Up") })
	reg.OnKey(tui.KeyDown, func() { events.Emit("Down") })
	reg.OnKey(tui.KeyLeft, func() { events.Emit("Left") })
	reg.OnKey(tui.KeyRight, func() { events.Emit("Right") })

	// Mount child components - collect their bindings
	childReg := NewRegistrar()
	counterComp.OnMount(childReg)
	reg.Merge(childReg)

	childReg = NewRegistrar()
	timerComp.OnMount(childReg)
	reg.Merge(childReg)

	childReg = NewRegistrar()
	togglesComp.OnMount(childReg)
	reg.Merge(childReg)

	return &interactiveApp{
		events:         events,
		registrar:      reg,
		counter:        counterComp,
		timer:          timerComp,
		toggles:        togglesComp,
		eventInspector: inspectorComp,
	}
}

// KeyMap returns merged key bindings from all components
func (a *interactiveApp) KeyMap() tui.KeyMap {
	return a.registrar.KeyMap()
}

// HandleMouse delegates to the registrar for automatic hit testing
func (a *interactiveApp) HandleMouse(me tui.MouseEvent) bool {
	return a.registrar.HandleMouse(me)
}

// tick calls all registered timer functions
func (a *interactiveApp) tick() {
	for _, t := range a.registrar.Timers() {
		t.fn()
	}
}

templ (a *interactiveApp) Render() {
	<div
		class="flex-col p-1 border-rounded gap-1"
		onTimer={tui.OnTimer(time.Second, a.tick)}>
		<div class="flex justify-between">
			<span class="text-gradient-cyan-magenta font-bold">{"Interactive Elements"}</span>
			<span class="font-dim">{"[q] quit"}</span>
		</div>
		<div class="flex gap-1">
			@a.Counter()
			@a.Timer()
		</div>
		<div class="flex gap-1">
			@a.Toggles()
			@a.EventInspector()
		</div>
	</div>
}
