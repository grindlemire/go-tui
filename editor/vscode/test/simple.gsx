// Simple GSX component example
// Tests basic syntax highlighting

package example

import (
	"github.com/grindlemire/go-tui/pkg/tui"
)

templ Header(title string) {
	<div border={tui.BorderSingle} padding={1}>
		<span>{title}</span>
	</div>
}

templ Footer() {
	<div padding={1}>
		<span>Footer content</span>
	</div>
}

templ SimpleCard(title string, content string) {
	<div border={tui.BorderRounded}>
		<span class="font-bold">{title}</span>
		<span>{content}</span>
	</div>
}
