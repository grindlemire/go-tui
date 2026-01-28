package main

import (
	"fmt"
)

// Header displays the title bar with instructions
func Header() Element {
	<div class="border-blue" border={tui.BorderSingle} height={3} direction={layout.Row} justify={layout.JustifyCenter} align={layout.AlignCenter}>
		<span class="font-bold text-white">{"Streaming DSL Demo - Use j/k to scroll, q to quit"}</span>
	</div>
}

// Footer displays status information
func Footer(lineCount int, elapsed int) Element {
	<div class="border-blue" border={tui.BorderSingle} height={3} direction={layout.Row} justify={layout.JustifyCenter} align={layout.AlignCenter}>
		<span class="text-white">{fmt.Sprintf("Lines: %d | Elapsed: %ds | Press q to exit", lineCount, elapsed)}</span>
	</div>
}

// StreamContent displays the streaming content area
@component StreamContent() {
	<div #Content
		class="flex-col border-cyan"
		border={tui.BorderSingle}
		flexGrow={1}
		scrollable={element.ScrollVertical}
		focusable={true}
		onKeyPress={handleScrollKeys}>
	</div>
}

// StreamApp is the main application component
@component StreamApp(lineCount int, elapsed int) {
	<div class="flex-col">
		@Header()
		@StreamContent()
		@Footer(lineCount, elapsed)
	</div>
}

// handleScrollKeys handles keyboard navigation for scrolling
// Note: App-level keys like quit are handled via SetGlobalKeyHandler
func handleScrollKeys(e tui.KeyEvent) {
	// This is a simple handler - the actual scrolling is done in main.go
	// because we need access to the Content element reference
}
