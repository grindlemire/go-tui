package main

import (
	"fmt"
	tui "github.com/grindlemire/go-tui"
)

// RefsDemo demonstrates named element references:
// - Simple refs: direct element access (#Header, #Content, #StatusBar)
// - Loop refs: slice of elements (#Items)
// - Conditional refs: may be nil (#Warning)
templ RefsDemo(items []string, showWarning bool, selectedIdx int) {
	<div class="flex-col" height={24} width={80}>
		<div
			#Header
			class="border p-1"
			height={3}
			direction={tui.Row}
			justify={tui.JustifyCenter}
			align={tui.AlignCenter}>
			<span class="font-bold text-cyan">{"Named Element Refs Demo"}</span>
		</div>
		<div
			#Content
			class="flex-col border p-1"
			flexGrow={1}
			scrollable={tui.ScrollVertical}
			direction={tui.Column}>
			<span class="font-bold text-white">{"Items (loop refs) - j/k to scroll, +/- to select"}</span>
			@for i, item := range items {
				<span #Items class={itemStyle(i, selectedIdx)}>{item}</span>
			}
		</div>
		@if showWarning {
			<div
				#Warning
				class="border-double p-1 text-yellow"
				height={3}
				direction={tui.Row}
				justify={tui.JustifyCenter}
				align={tui.AlignCenter}>
				<span class="font-bold">
					{"⚠ Warning: This is a conditional ref (may be nil)"}
				</span>
			</div>
		}
		<div
			#StatusBar
			class="border p-1"
			height={3}
			direction={tui.Row}
			justify={tui.JustifySpaceBetween}
			align={tui.AlignCenter}>
			<span class="text-white">
				{"j/k: scroll | +/-: select | Tab: warning | d: switch demo | q: quit"}
			</span>
			<span class="font-dim">{fmt.Sprintf(" Selected: %d", selectedIdx)}</span>
		</div>
	</div>
}

// KeyedRefsDemo demonstrates keyed refs that generate map[KeyType]*tui.Element
// Use key={expr} inside @for loops for stable key-based element access
templ KeyedRefsDemo(users []User) {
	<div class="flex-col p-1" height={20} width={60}>
		<div
			class="border p-1"
			height={3}
			direction={tui.Row}
			justify={tui.JustifyCenter}
			align={tui.AlignCenter}>
			<span class="font-bold text-cyan">{"Keyed Refs Demo (map access)"}</span>
		</div>
		<div class="flex-col border p-1" flexGrow={1}>
			<span class="font-bold text-white">{"Users (keyed by ID)"}</span>
			@for _, user := range users {
				<span #Users>{fmt.Sprintf("[%s] %s", user.ID, user.Name)}</span>
			}
		</div>
		<div class="border p-1" height={2}>
			<span class="font-dim">{"1-3: highlight user | d: switch demo | q: quit"}</span>
		</div>
	</div>
}

func itemStyle(idx, selected int) string {
	if idx == selected {
		return "font-bold text-cyan"
	}
	return "text-white"
}
