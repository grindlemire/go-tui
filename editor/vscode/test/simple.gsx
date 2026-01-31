// Simple GSX component example
// Tests basic syntax highlighting

package example

import (
	"fmt"
	"github.com/grindlemire/go-tui/pkg/tui"
)

templ Header(title string) {
	<div class="border-single p-1">
		<span class="font-bold">{title}</span>
	</div>
}

templ Footer() {
	<div class="p-1">
		<span>Footer content</span>
	</div>
}

templ SimpleCard(title string, content string) {
	<div class="border-rounded">
		<span class="font-bold">{title}</span>
		<span>{content}</span>
	</div>
}

// Named ref example
templ Layout(title string) {
	<div #Main class="flex-col gap-1">
		<span #Title class="font-bold">{title}</span>
		<span>Body content</span>
	</div>
}

// State variable example
templ Counter() {
	count := tui.NewState(0)
	<div class="flex-col gap-1 p-1 border-single">
		<span>{fmt.Sprintf("Count: %d", count.Get())}</span>
	</div>
}
