package main

import tui "github.com/grindlemire/go-tui"

// Card is a reusable component that accepts children
templ Card(title string) {
	<div class="border-rounded p-1 flex-col">
		<span class="text-gradient-cyan-magenta font-bold">{title}</span>
		<hr class="border" />
		{children...}
	</div>
}

// Badge is a simple styled component
templ Badge(text string) {
	<span class="bg-gradient-blue-cyan text-white font-bold">{" " + text + " "}</span>
}

// Header shows a component without children
templ Header(text string) {
	<div class="border-double p-1">
		<span class="text-gradient-blue-cyan font-bold">{text}</span>
	</div>
}

// StatusLine pairs a dim label with a bright value
templ StatusLine(label string, value string) {
	<div class="flex gap-1">
		<span class="font-dim">{label}</span>
		<span class="text-cyan font-bold">{value}</span>
	</div>
}

type compositionApp struct{}

func App() *compositionApp {
	return &compositionApp{}
}

func (a *compositionApp) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnRune('q', func(ke tui.KeyEvent) { tui.Stop() }),
		tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { tui.Stop() }),
	}
}

// App composes the other components together
templ (a *compositionApp) Render() {
	<div class="flex-col p-1 border-rounded gap-1">
		<div class="flex justify-between">
			<span class="text-gradient-cyan-magenta font-bold">{"Component Composition"}</span>
			<span class="text-blue font-bold">{"Nested Components"}</span>
		</div>
		<div class="flex gap-1">
			<div class="border-single p-1 flex-col" flexGrow={1.0}>
				<span class="font-bold">{"@Component (leaf)"}</span>
				@Header("go-tui")
				@Badge("Framework")
			</div>
			<div class="border-single p-1 flex-col" flexGrow={1.0}>
				<span class="font-bold">{"@Component {children}"}</span>
				@Card("User Profile") {
					@StatusLine("Name:", "Alice")
					@StatusLine("Role:", "Admin")
					<div class="flex gap-1">
						<span class="font-dim">Status:</span>
						@Badge("Active")
					</div>
				}
			</div>
			<div class="border-single p-1 flex-col" flexGrow={1.0}>
				<span class="font-bold">{"Deep Nesting"}</span>
				@Card("Settings") {
					@StatusLine("Theme:", "Dark")
					@StatusLine("Notify:", "On")
					<div class="flex gap-1">
						<span class="font-dim">Tags:</span>
						@Badge("New")
						@Badge("v1.0")
					</div>
				}
			</div>
		</div>
		<div class="flex justify-center">
			<span class="font-dim">{"[q] quit"}</span>
		</div>
	</div>
}
