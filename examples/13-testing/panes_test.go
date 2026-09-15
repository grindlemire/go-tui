package main

import (
	"strings"
	"testing"

	tui "github.com/grindlemire/go-tui"
)

// The generated Render only forwards app to the element expressions, and an
// Element's Render(app) ignores it, so the component renders without an App.
func TestPanes_RendersActiveElementExpression(t *testing.T) {
	type tc struct {
		active      int
		wantVisible []string
		wantHidden  string
	}

	tests := map[string]tc{
		"first pane active": {
			active:      0,
			wantVisible: []string{"Pane A", "item 1", "item 2", "footer"},
			wantHidden:  "Pane B",
		},
		"second pane active": {
			active:      1,
			wantVisible: []string{"Pane B", "item 1", "item 2", "footer"},
			wantHidden:  "Pane A",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			p := NewPanes(tt.active)
			root := p.Render(nil)

			if root.Children()[0] != p.content[tt.active] {
				t.Fatalf("first child is not the active pane element")
			}

			buf := tui.NewBuffer(30, 6)
			term := tui.NewMockTerminal(30, 6)
			root.RenderTo(buf, 30, 6)
			tui.Render(term, buf)

			output := term.StringTrimmed()
			for _, want := range tt.wantVisible {
				if !strings.Contains(output, want) {
					t.Errorf("expected %q in output:\n%s", want, output)
				}
			}
			if strings.Contains(output, tt.wantHidden) {
				t.Errorf("did not expect %q in output:\n%s", tt.wantHidden, output)
			}
		})
	}
}
