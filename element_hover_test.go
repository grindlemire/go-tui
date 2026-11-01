package tui

import "testing"

func TestHoverClass_ApplyAndRestore(t *testing.T) {
	type tc struct {
		opts      []Option
		wantHover Style // text style while hovered
		wantBg    *Style
		wantAfter Style // text style after unhover
	}

	blueBg := NewStyle().Background(Blue)
	tests := map[string]tc{
		"background only": {
			opts:   []Option{WithClass("px-1 hover:bg-blue")},
			wantBg: &blueBg,
		},
		"text style layers onto base": {
			opts:      []Option{WithClass("text-white hover:font-bold")},
			wantHover: NewStyle().Foreground(White).Bold(),
			wantBg:    nil,
			wantAfter: NewStyle().Foreground(White),
		},
		"hover color overrides base color while hovered": {
			opts:      []Option{WithTextStyle(NewStyle().Foreground(White)), WithHoverClass("text-red")},
			wantHover: NewStyle().Foreground(Red),
			wantAfter: NewStyle().Foreground(White),
		},
		"layout classes are ignored": {
			opts:   []Option{WithHoverClass("p-3 bg-blue")},
			wantBg: &blueBg,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			e := New(tt.opts...)
			before := *e
			e.setHoverStyle(true)
			if got, _ := effectiveStyles(e, inheritedStyle{}); got != tt.wantHover {
				t.Errorf("hovered text style = %+v, want %+v", got, tt.wantHover)
			}
			if (e.background == nil) != (tt.wantBg == nil) || (e.background != nil && *e.background != *tt.wantBg) {
				t.Errorf("hovered background = %v, want %v", e.background, tt.wantBg)
			}
			if e.style.Padding != before.style.Padding {
				t.Errorf("hover changed padding: %+v", e.style.Padding)
			}

			e.setHoverStyle(false)
			if e.textStyle != tt.wantAfter {
				t.Errorf("unhovered text style = %+v, want %+v", e.textStyle, tt.wantAfter)
			}
			if e.background != before.background || e.textStyleSet != before.textStyleSet {
				t.Errorf("unhover did not restore: background %v (was %v), textStyleSet %v (was %v)",
					e.background, before.background, e.textStyleSet, before.textStyleSet)
			}
		})
	}
}

func TestHoverClass_SetterAfterHoverWins(t *testing.T) {
	e := New(WithHoverClass("bg-blue"))
	e.setHoverStyle(true)
	red := NewStyle().Background(Red)
	e.Apply(WithBackground(red))
	e.setHoverStyle(false)
	if e.background == nil || *e.background != red {
		t.Errorf("unhover clobbered a background set while hovered: %v", e.background)
	}
}

func TestHoverClass_SetClassWhileHovered(t *testing.T) {
	e := New(WithClass("hover:bg-blue"))
	e.setHoverStyle(true)

	e.SetClass("hover:bg-red")
	red := NewStyle().Background(Red)
	if !e.hoverActive || e.background == nil || *e.background != red {
		t.Fatalf("after SetClass while hovered: active %v, background %v, want red", e.hoverActive, e.background)
	}

	e.SetClass("p-1")
	if e.hoverActive || e.background != nil || e.hoverClass != "" {
		t.Errorf("SetClass without hover classes left hover state: active %v, background %v, class %q",
			e.hoverActive, e.background, e.hoverClass)
	}
}

func TestHoverClass_TextStyleLayersOnInheritedStyle(t *testing.T) {
	type tc struct {
		hoverClass string
		wantFg     Color
		wantBold   bool
	}

	tests := map[string]tc{
		// The container sets no text style of its own; hovering adds bold on
		// top of the cyan it inherits instead of replacing it.
		"attribute keeps inherited color": {hoverClass: "font-bold", wantFg: Cyan, wantBold: true},
		"color overrides inherited color": {hoverClass: "text-red", wantFg: Red},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			outer := New(WithTextStyle(NewStyle().Foreground(Cyan)), WithWidth(10), WithHeight(1))
			container := New(WithHoverClass(tt.hoverClass), WithWidth(10), WithHeight(1))
			container.AddChild(New(WithText("child")))
			outer.AddChild(container)
			container.setHoverStyle(true)

			buf := NewBuffer(10, 1)
			outer.RenderTo(buf, 10, 1)
			got := buf.Cell(0, 0).Style
			if got.Fg != tt.wantFg || got.HasAttr(AttrBold) != tt.wantBold {
				t.Errorf("hovered child text style = %+v, want fg %v bold %v", got, tt.wantFg, tt.wantBold)
			}

			container.setHoverStyle(false)
			outer.RenderTo(buf, 10, 1)
			if got := buf.Cell(0, 0).Style; got.Fg != Cyan || got.HasAttr(AttrBold) {
				t.Errorf("unhovered child text style = %+v, want plain cyan", got)
			}
		})
	}
}
