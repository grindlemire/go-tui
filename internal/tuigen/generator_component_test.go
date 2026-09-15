package tuigen

import (
	"slices"
	"testing"
)

func TestGenerator_TrackComponentExprField(t *testing.T) {
	type tc struct {
		expr string
		want []string
	}

	tests := map[string]tc{
		"plain field":    {expr: "c.footer", want: []string{"footer"}},
		"indexed field":  {expr: "c.content[c.active]", want: nil},
		"literal index":  {expr: "c.content[0]", want: nil},
		"nested field":   {expr: "c.a.b", want: nil},
		"method call":    {expr: "c.view()", want: nil},
		"other receiver": {expr: "x.footer", want: nil},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			g := &Generator{currentReceiver: "c"}
			g.trackComponentExprField(tt.expr)
			if !slices.Equal(g.componentExprFields, tt.want) {
				t.Errorf("componentExprFields = %v, want %v", g.componentExprFields, tt.want)
			}
		})
	}
}
