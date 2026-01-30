package main

templ Conditionals(enabled bool) {
	<div class="flex-col gap-1 p-2 border-rounded">
		<span class="font-bold">Conditional Rendering</span>
		<hr class="border" />

		<div class="flex gap-1">
			<span>Status:</span>
			@if enabled {
				<span class="text-green font-bold">Enabled</span>
			} @else {
				<span class="text-red font-bold">Disabled</span>
			}
		</div>

		@if enabled {
			<div class="border-single p-1 bg-green">
				<span>This content only shows when enabled!</span>
			</div>
		}

		<br />
		<span class="font-dim">Press t to toggle, q to quit</span>
	</div>
}
