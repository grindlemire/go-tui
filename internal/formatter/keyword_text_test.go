package formatter

import "testing"

// The formatter used to drop Go keywords from element text ("Press Escape to
// return" came back as "Press Escape to"), so already formatted prose must
// survive a pass unchanged.
func TestFormat_KeywordsInText(t *testing.T) {
	input := `package test

import tui "github.com/grindlemire/go-tui"

templ Hint(show bool) {
	<div class="flex-col">
		<span>Press Escape to return</span>
		<span>wait for it</span>
		<span>return</span>
		<span>or else</span>
		if show {
			<span>for the win</span>
		}
	</div>
}
`
	got, err := New().Format("test.gsx", input)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}
	if got != input {
		t.Errorf("Format() changed keyword prose:\n--- want\n%s\n--- got\n%s", input, got)
	}
}
