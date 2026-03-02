package main

import (
	"fmt"
	tui "github.com/grindlemire/go-tui"
)

templ BuildReport(project string, status string, duration string, tests int, passed int) {
	<div class="flex-row justify-center">
		<div class="flex-col border-rounded border-cyan p-1 w-1/2">
			<div class="flex-row justify-between">
				<span class="font-bold text-cyan">{project}</span>
				<span class="font-bold text-green">{status}</span>
			</div>
			<hr />
			<div class="flex-row gap-4">
				<span class="text-dim">Duration: {duration}</span>
				<span class="text-dim">Tests: {fmt.Sprintf("%d/%d passed", passed, tests)}</span>
			</div>
		</div>
	</div>
}
