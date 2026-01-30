package main

// Card is a reusable component that accepts children
templ Card(title string) {
	<div class="border-rounded p-1 flex-col">
		<span class="font-bold text-cyan">{title}</span>
		<hr class="border" />
		{children...}
	</div>
}

// Badge is a simple styled component
templ Badge(text string) {
	<span class="bg-blue text-white">{" " + text + " "}</span>
}

// Header shows a component without children
templ Header(text string) {
	<div class="border-double p-1">
		<span class="font-bold">{text}</span>
	</div>
}

// App composes the other components together
templ App() {
	<div class="flex-col gap-2 p-2">
		@Header("Component Composition Demo")

		@Card("User Profile") {
			<span>Name: Alice</span>
			<span>Role: Admin</span>
			<div class="flex gap-1">
				<span>Status:</span>
				@Badge("Active")
			</div>
		}

		@Card("Settings") {
			<span>Theme: Dark</span>
			<span>Notifications: On</span>
			<div class="flex gap-1">
				<span>Version:</span>
				@Badge("v1.0")
			</div>
		}

		<span class="font-dim">Press q to quit</span>
	</div>
}
