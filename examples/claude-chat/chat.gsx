package main

import tui "github.com/grindlemire/go-tui"

// InputBox renders a single-line input box pinned to the bottom of the inline region.
// The height parameter should match the app's WithInlineHeight value.
templ InputBox(text string, height int) {
	<div class="flex-col justify-end" height={height}>
		<div class="border-rounded p-1">
			<span>{"> " + text + "\u2588"}</span>
		</div>
	</div>
}
