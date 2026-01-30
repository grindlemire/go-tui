package main

import "fmt"

templ Loops(items []string, selected int) {
	<div class="flex-col gap-1 p-2 border-rounded">
		<span class="font-bold">Loop Rendering</span>
		<hr class="border" />

		<div class="flex-col">
			@for i, item := range items {
				@if i == selected {
					<span class="bg-blue text-white">{fmt.Sprintf("> %d. %s", i+1, item)}</span>
				} @else {
					<span>{fmt.Sprintf("  %d. %s", i+1, item)}</span>
				}
			}
		</div>

		<br />
		<span class="font-dim">Press j/k to move, q to quit</span>
	</div>
}
