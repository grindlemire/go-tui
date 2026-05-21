package main

import (
	"fmt"

	tui "github.com/grindlemire/go-tui"
)

// maxEventLog is intentionally generous: with a small cap the log becomes a
// sliding window, and at scrollY=0 the user sees the oldest *currently
// stored* entries — which keep changing as new events push old ones out, so
// it looks like the viewport isn't scrolling when it actually is.
const maxEventLog = 500

type eventDump struct {
	log     *tui.State[[]string]
	scrollY *tui.State[int]
	// following is wrapped in State[bool] so the generated UpdateProps treats
	// it as framework-managed and does NOT overwrite it from a freshly-
	// constructed eventDump every render. A plain bool field would get
	// reset to its constructor value on every re-render, causing scroll
	// position to snap back to the bottom.
	following *tui.State[bool]
	// seq is a monotonically increasing counter prefixed onto every log line
	// so the entries are visually distinct even when the same key is pressed
	// repeatedly. Wrapped in State so the generator treats it as framework-
	// managed and doesn't reset it across renders.
	seq  *tui.State[int]
	feed *tui.Ref
}

// EventDump constructs a diagnostic component that logs every key and mouse
// event it receives, so a human can eyeball whether the event pipeline is
// behaving correctly. Resize is exercised implicitly: the layout reflows on
// every WINDOW_BUFFER_SIZE_EVENT, so dragging the terminal corner is enough
// to verify it.
func EventDump() *eventDump {
	return &eventDump{
		log:       tui.NewState([]string{}),
		scrollY:   tui.NewState(0),
		following: tui.NewState(true),
		seq:       tui.NewState(0),
		feed:      tui.NewRef(),
	}
}

func (e *eventDump) Watchers() []tui.Watcher {
	return []tui.Watcher{
		tui.OnChange(e.log, e.autoScrollToBottom),
	}
}

func (e *eventDump) autoScrollToBottom(_ []string) {
	if !e.following.Get() {
		return
	}
	el := e.feed.El()
	if el == nil {
		return
	}
	_, maxY := el.MaxScroll()
	e.scrollY.Set(maxY)
}

// scrollBy adjusts the log viewport by delta rows, clamped to [0, maxY], and
// toggles the follow flag based on whether the user has scrolled away from
// the bottom. delta < 0 scrolls up; delta > 0 scrolls down.
func (e *eventDump) scrollBy(delta int) {
	el := e.feed.El()
	if el == nil {
		return
	}
	_, maxY := el.MaxScroll()
	newY := e.scrollY.Get() + delta
	if newY < 0 {
		newY = 0
	}
	if newY > maxY {
		newY = maxY
	}
	e.scrollY.Set(newY)
	e.following.Set(newY >= maxY)
}

func (e *eventDump) append(line string) {
	n := e.seq.Get() + 1
	e.seq.Set(n)
	tagged := fmt.Sprintf("#%03d  %s", n, line)
	e.log.Update(func(prev []string) []string {
		next := append(prev, tagged)
		if len(next) > maxEventLog {
			next = next[len(next)-maxEventLog:]
		}
		return next
	})
}

func (e *eventDump) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		// High-priority quit handlers stop propagation so the AnyKey catch-all
		// below doesn't double-log the quit keystroke.
		tui.OnStop(tui.KeyEscape, func(ke tui.KeyEvent) {
			e.append(formatKey(ke) + "  [QUIT]")
			ke.App().Stop()
		}),
		tui.OnStop(tui.Rune('q'), func(ke tui.KeyEvent) {
			e.append(formatKey(ke) + "  [QUIT]")
			ke.App().Stop()
		}),
		tui.On(tui.AnyKey, func(ke tui.KeyEvent) {
			e.append(formatKey(ke))
		}),
	}
}

// statusLine reports both the State view of scroll position and the actual
// Element view from the previous render, so we can tell whether the State
// binding is propagating to the rendered viewport.
func (e *eventDump) statusLine() string {
	stateY := e.scrollY.Get()
	following := e.following.Get()
	entries := len(e.log.Get())
	elY, contentH, viewportH := -1, -1, -1
	if el := e.feed.El(); el != nil {
		_, elY = el.ScrollOffset()
		_, contentH = el.ContentSize()
		_, viewportH = el.ViewportSize()
	}
	return fmt.Sprintf(
		"STATUS: stateY=%d elY=%d  following=%v  entries=%d  contentH=%d viewportH=%d",
		stateY, elY, following, entries, contentH, viewportH,
	)
}

// HandleMouse logs every mouse event for diagnostic purposes and also wires
// the wheel to scroll the log. Returns false on non-wheel events so element
// hit-tests still see them.
func (e *eventDump) HandleMouse(me tui.MouseEvent) bool {
	e.append(formatMouse(me))
	switch me.Button {
	case tui.MouseWheelUp:
		e.scrollBy(-3)
		return true
	case tui.MouseWheelDown:
		e.scrollBy(3)
		return true
	}
	return false
}

func formatKey(ke tui.KeyEvent) string {
	if ke.Key == tui.KeyRune {
		return fmt.Sprintf("KEY  Rune=%q  Mod=%s", ke.Rune, ke.Mod)
	}
	return fmt.Sprintf("KEY  %s  Mod=%s", ke.Key, ke.Mod)
}

func formatMouse(me tui.MouseEvent) string {
	return fmt.Sprintf("MOUSE  %s %s  X=%d Y=%d  Mod=%s",
		buttonString(me.Button), actionString(me.Action), me.X, me.Y, me.Mod)
}

func buttonString(b tui.MouseButton) string {
	switch b {
	case tui.MouseLeft:
		return "Left"
	case tui.MouseMiddle:
		return "Middle"
	case tui.MouseRight:
		return "Right"
	case tui.MouseWheelUp:
		return "WheelUp"
	case tui.MouseWheelDown:
		return "WheelDown"
	case tui.MouseNone:
		return "None"
	}
	return "?"
}

func actionString(a tui.MouseAction) string {
	switch a {
	case tui.MousePress:
		return "Press"
	case tui.MouseRelease:
		return "Release"
	case tui.MouseDrag:
		return "Drag"
	}
	return "?"
}

templ (e *eventDump) Render() {
	<div class="flex-col gap-1 p-1 border-rounded border-cyan grow">
		<span class="text-gradient-cyan-magenta font-bold">Event Dump</span>
		<span class="text-yellow font-bold">{e.statusLine()}</span>
		<span class="font-dim">{"Press any key or click/scroll the mouse. Esc or q to quit."}</span>
		<hr class="border-single" />
		<div
			ref={e.feed}
			class="flex-col grow overflow-y-scroll scrollbar-hidden"
			scrollOffset={0, e.scrollY.Get()}>
			for _, line := range e.log.Get() {
				<span class="truncate">{line}</span>
			}
		</div>
	</div>
}
