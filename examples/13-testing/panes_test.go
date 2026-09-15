package main

import (
	"strings"
	"testing"

	tui "github.com/grindlemire/go-tui"
)

// Element.Render ignores app, so nil is fine here.
func TestPanes_RendersActiveElementExpression(t *testing.T) {
	type tc struct {
		active      int
		wantVisible []string
		wantHidden  string
	}

	tests := map[string]tc{
		"first pane active": {
			active:      0,
			wantVisible: []string{"Pane A", "footer"},
			wantHidden:  "Pane B",
		},
		"second pane active": {
			active:      1,
			wantVisible: []string{"Pane B", "footer"},
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

type widget struct {
	bound   bool
	unbound bool
}

func (w *widget) Render(app *tui.App) *tui.Element {
	return tui.New(tui.WithText("widget"))
}

func (w *widget) BindApp(app *tui.App) { w.bound = true }

func (w *widget) UnbindApp() { w.unbound = true }

// The generated BindApp and UnbindApp range over the widgets slice because the
// template renders it with a for loop, so each component sees both calls.
func TestPanes_ForwardsBindAppToSliceComponents(t *testing.T) {
	w := &widget{}
	p := NewPanes(0, w)

	p.BindApp(nil)
	if !w.bound {
		t.Fatal("BindApp was not forwarded to the widget in p.widgets")
	}

	p.UnbindApp()
	if !w.unbound {
		t.Fatal("UnbindApp was not forwarded to the widget in p.widgets")
	}
}
