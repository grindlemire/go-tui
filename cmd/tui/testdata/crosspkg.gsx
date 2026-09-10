package testdata

import (
	"example.com/external/widgets"
	tui "github.com/grindlemire/go-tui"
)

templ CrossPackage(title string) {
	<div>
		@widgets.Header(title)
	</div>
}
