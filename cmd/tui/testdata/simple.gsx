package testdata

import tui "github.com/grindlemire/go-tui"

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
