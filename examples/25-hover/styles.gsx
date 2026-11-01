package main

import tui "github.com/grindlemire/go-tui"

// stylesPage shows how hover text styles cascade to children and layer onto
// inherited styles, plus gradient hover classes.
type stylesPage struct {
	d *hoverDemo
}

templ (p *stylesPage) Render() {
	<div class="flex gap-2 w-full h-full">
		<div class="flex-col border-rounded px-1 w-28 text-white hover:text-bright-yellow" height={7} onMouseEnter={p.d.logged("enter cascade box")} onMouseLeave={p.d.logged("leave cascade box")}>
			<span class="font-bold">Cascade</span>
			<span>hover:text-bright-yellow</span>
			<span>reaches every child</span>
			<span>without its own color,</span>
			<span class="text-cyan">but not this cyan one.</span>
		</div>
		<div class="flex-col text-cyan">
			<span>Inherited cyan:</span>
			<div class="flex-col border-rounded px-1 w-26 hover:font-bold hover:underline" height={5} onMouseEnter={p.d.logged("enter layered box")} onMouseLeave={p.d.logged("leave layered box")}>
				<span>hover:font-bold</span>
				<span>hover:underline</span>
				<span>layer on, cyan stays.</span>
			</div>
		</div>
		<div class="flex-col gap-1">
			<div class="flex-col border-rounded px-1 w-26 hover:border-gradient-cyan-magenta" height={4} onMouseEnter={p.d.logged("enter gradient box")} onMouseLeave={p.d.logged("leave gradient box")}>
				<span class="hover:text-gradient-yellow-red">Gradient text</span>
				<span class="font-dim">and gradient border</span>
			</div>
			<span class="px-2 border-single w-26 hover:bg-gradient-blue-magenta hover:text-bright-white" onMouseEnter={p.d.logged("enter gradient button")} onMouseLeave={p.d.logged("leave gradient button")}>Gradient background</span>
		</div>
	</div>
}
