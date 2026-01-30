package main

templ Layout() {
	<div class="flex-col gap-1 p-1">
		<span class="font-bold">Flexbox Layout Demo</span>
		<hr class="border" />

		<div class="flex-col gap-1">
			<span class="font-dim">Row layout (flex):</span>
			<div class="flex gap-1">
				<span class="border-single p-1">One</span>
				<span class="border-single p-1">Two</span>
				<span class="border-single p-1">Three</span>
			</div>
		</div>

		<div class="flex-col gap-1">
			<span class="font-dim">Column layout (flex-col):</span>
			<div class="flex-col border-single">
				<span class="p-1">First</span>
				<span class="p-1">Second</span>
				<span class="p-1">Third</span>
			</div>
		</div>

		<div class="flex-col gap-1">
			<span class="font-dim">Justify content:</span>
			<div class="flex gap-1">
				<div class="flex justify-start border-single w-20">
					<span>start</span>
				</div>
				<div class="flex justify-center border-single w-20">
					<span>center</span>
				</div>
				<div class="flex justify-end border-single w-20">
					<span>end</span>
				</div>
			</div>
		</div>

		<div class="flex-col gap-1">
			<span class="font-dim">Flex grow:</span>
			<div class="flex gap-1 w-60">
				<span class="border-single p-1">Fixed</span>
				<span class="border-single p-1 flex-grow">Grows to fill space</span>
				<span class="border-single p-1">Fixed</span>
			</div>
		</div>

		<span class="font-dim">Press q to quit</span>
	</div>
}
