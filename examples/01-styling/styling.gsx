package main

templ Styling() {
	<div class="flex-col gap-1 p-2 border-rounded">
		<span class="font-bold">Text Styles</span>
		<hr class="border" />
		<div class="flex-col gap-1">
			<span class="font-bold">Bold text</span>
			<span class="font-dim">Dim text</span>
			<span class="italic">Italic text</span>
			<span class="underline">Underlined text</span>
		</div>

		<br />
		<span class="font-bold">Text Colors</span>
		<hr class="border" />
		<div class="flex gap-2">
			<span class="text-red">Red</span>
			<span class="text-green">Green</span>
			<span class="text-blue">Blue</span>
			<span class="text-cyan">Cyan</span>
			<span class="text-magenta">Magenta</span>
			<span class="text-yellow">Yellow</span>
		</div>

		<br />
		<span class="font-bold">Background Colors</span>
		<hr class="border" />
		<div class="flex gap-2">
			<span class="bg-red"> Red </span>
			<span class="bg-green"> Green </span>
			<span class="bg-blue"> Blue </span>
			<span class="bg-cyan"> Cyan </span>
			<span class="bg-magenta"> Magenta </span>
			<span class="bg-yellow"> Yellow </span>
		</div>

		<br />
		<span class="font-dim">Press q to quit</span>
	</div>
}
