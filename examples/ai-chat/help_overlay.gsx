package main

import tui "github.com/grindlemire/go-tui"

templ HelpOverlay() {
	<div class="flex justify-center items-center" flexGrow={1}>
		<div class="border-rounded border-cyan p-2" width={50}>
			<div class="flex-col gap-1">
				<span class="text-gradient-cyan-magenta font-bold text-center">{"Keyboard Shortcuts"}</span>
				<hr />
				<div class="flex justify-between">
					<span class="font-bold">{"Ctrl+,"}</span>
					<span>{"Open settings"}</span>
				</div>
				<div class="flex justify-between">
					<span class="font-bold">{"Ctrl+?"}</span>
					<span>{"Toggle this help"}</span>
				</div>
				<div class="flex justify-between">
					<span class="font-bold">{"Ctrl+L"}</span>
					<span>{"Clear conversation"}</span>
				</div>
				<div class="flex justify-between">
					<span class="font-bold">{"Ctrl+C"}</span>
					<span>{"Cancel/Quit"}</span>
				</div>
				<hr />
				<span class="font-dim text-center">{"Message Navigation"}</span>
				<div class="flex justify-between">
					<span class="font-bold">{"j/k"}</span>
					<span>{"Move down/up"}</span>
				</div>
				<div class="flex justify-between">
					<span class="font-bold">{"g/G"}</span>
					<span>{"First/Last message"}</span>
				</div>
				<div class="flex justify-between">
					<span class="font-bold">{"c"}</span>
					<span>{"Copy message"}</span>
				</div>
				<div class="flex justify-between">
					<span class="font-bold">{"r"}</span>
					<span>{"Retry response"}</span>
				</div>
				<hr />
				<span class="font-dim text-center">{"Press any key to close"}</span>
			</div>
		</div>
	</div>
}
