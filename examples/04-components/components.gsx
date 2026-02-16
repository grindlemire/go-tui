package main

import tui "github.com/grindlemire/go-tui"

type componentsApp struct {
	scrollY *tui.State[int]
	content *tui.Ref
}

func App() *componentsApp {
	return &componentsApp{
		scrollY: tui.NewState(0),
		content: tui.NewRef(),
	}
}

func (a *componentsApp) scrollBy(delta int) {
	el := a.content.El()
	if el == nil {
		return
	}
	_, maxY := el.MaxScroll()
	newY := a.scrollY.Get() + delta
	if newY < 0 {
		newY = 0
	} else if newY > maxY {
		newY = maxY
	}
	a.scrollY.Set(newY)
}

func (a *componentsApp) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnRune('q', func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnRune('j', func(ke tui.KeyEvent) { a.scrollBy(1) }),
		tui.OnRune('k', func(ke tui.KeyEvent) { a.scrollBy(-1) }),
		tui.OnKey(tui.KeyDown, func(ke tui.KeyEvent) { a.scrollBy(1) }),
		tui.OnKey(tui.KeyUp, func(ke tui.KeyEvent) { a.scrollBy(-1) }),
	}
}

func (a *componentsApp) HandleMouse(me tui.MouseEvent) bool {
	switch me.Button {
	case tui.MouseWheelUp:
		a.scrollBy(-1)
		return true
	case tui.MouseWheelDown:
		a.scrollBy(1)
		return true
	}
	return false
}

templ (a *componentsApp) Render() {
	<div class="flex-col p-2 gap-2">
		@Header("Component Showcase")
		// User Cards row
		<div class="flex gap-2">
			@UserCard("Alice", "Engineer", true)
			@UserCard("Bob", "Designer", false)
			@UserCard("Carol", "Manager", true)
		</div>

		// Card with children
		<div class="flex gap-2">
			@Card("System Info") {
				@StatusLine("Version:", "1.2.0")
				@StatusLine("Uptime:", "3d 14h")
				@StatusLine("Memory:", "1.2 GB")
			}
			@Card("Configuration") {
				@StatusLine("Theme:", "Dark")
				@StatusLine("Notify:", "On")
				<div class="flex gap-1">
					<span class="font-dim">Tags:</span>
					@Badge("New", "text-green")
					@Badge("v1.0", "text-cyan")
				</div>
			}
		</div>

		// Status Bar
		@StatusBar()
		<div class="flex justify-center">
			<span class="font-dim">j/k scroll|q to quit</span>
		</div>
	</div>
}

// Card wraps child elements in a titled, bordered container
templ Card(title string) {
	<div class="border-rounded p-1 flex-col gap-1 w-full" flexGrow={1.0}>
		<span class="text-gradient-cyan-magenta font-bold">{title}</span>
		<hr class="border-single" />
		<div class="flex-row w-full justify-between">
			{children...}
		</div>
	</div>
}

// Badge renders a styled inline label
templ Badge(label string, color string) {
	<span class={color + " font-bold px-1"}>{label}</span>
}

// Header renders a gradient-bordered title
templ Header(title string) {
	<div class="border-rounded border-gradient-cyan-magenta p-1 flex justify-center">
		<span class="text-gradient-cyan-magenta font-bold">{title}</span>
	</div>
}

// StatusLine pairs a dim label with a highlighted value
templ StatusLine(label string, value string) {
	<div class="flex gap-1">
		<span class="font-dim">{label}</span>
		<span class="text-cyan font-bold">{value}</span>
	</div>
}

func statusLabel(online bool) string {
	if online {
		return "Online"
	}
	return "Offline"
}

func statusColor(online bool) string {
	if online {
		return "text-green"
	}
	return "font-dim"
}

// UserCard displays a user profile card
templ UserCard(name string, role string, online bool) {
	@Card(name) {
		<span class="font-dim">{role}</span>
		@Badge(statusLabel(online), statusColor(online))
	}
}

// StatusBar renders a horizontal status bar with multiple items
templ StatusBar() {
	<div class="border-rounded p-1 flex gap-2 justify-center">
		@Badge("Build passed", "text-green")
		<span class="font-dim">|</span>
		@Badge("3 warnings", "text-yellow")
		<span class="font-dim">|</span>
		@Badge("v1.2.0", "text-cyan")
	</div>
}
