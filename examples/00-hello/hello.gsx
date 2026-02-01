package main

import tui "github.com/grindlemire/go-tui"

templ Hello() {
	<div class="flex-col items-center justify-center h-full">
		<span class="font-bold text-red">Hello, TUI!</span>
		<span class="font-dim">Press q to quit</span>
	</div>
}
