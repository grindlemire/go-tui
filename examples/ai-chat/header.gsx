package main

import (
	"fmt"
	tui "github.com/grindlemire/go-tui"
)

templ Header(state *AppState) {
	<div class="border-rounded p-1" height={3} direction={tui.Row} justify={tui.JustifySpaceBetween} align={tui.AlignCenter}>
		<span class="text-gradient-cyan-magenta font-bold">{"  AI Chat"}</span>
		<div class="flex gap-2">
			<span class="font-dim">{state.Model.Get()}</span>
			<span class="text-cyan">{fmt.Sprintf("%d tokens", state.TotalTokens.Get())}</span>
			<span class="font-dim">{"Ctrl+? help"}</span>
		</div>
	</div>
}
