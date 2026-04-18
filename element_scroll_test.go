package tui

import (
	"testing"
)

func TestElement_Scroll(t *testing.T) {
	type tc struct {
		setup func(e *Element)
		check func(t *testing.T, e *Element)
	}

	tests := map[string]tc{
		"default scroll mode is none": {
			setup: func(e *Element) {},
			check: func(t *testing.T, e *Element) {
				if e.ScrollModeValue() != ScrollNone {
					t.Errorf("got %v, want ScrollNone", e.ScrollModeValue())
				}
			},
		},
		"scrollable sets vertical scroll": {
			setup: func(e *Element) {
				WithScrollable(ScrollVertical)(e)
			},
			check: func(t *testing.T, e *Element) {
				if !e.IsScrollable() {
					t.Error("expected IsScrollable() = true")
				}
			},
		},
		"scroll offset starts at zero": {
			setup: func(e *Element) {},
			check: func(t *testing.T, e *Element) {
				x, y := e.ScrollOffset()
				if x != 0 || y != 0 {
					t.Errorf("scroll offset = (%d, %d), want (0, 0)", x, y)
				}
			},
		},
		"ScrollMode vertical": {
			setup: func(e *Element) {
				WithScrollable(ScrollVertical)(e)
			},
			check: func(t *testing.T, e *Element) {
				if e.ScrollModeValue() != ScrollVertical {
					t.Errorf("got %v, want ScrollVertical", e.ScrollModeValue())
				}
			},
		},
		"ScrollMode horizontal": {
			setup: func(e *Element) {
				WithScrollable(ScrollHorizontal)(e)
			},
			check: func(t *testing.T, e *Element) {
				if e.ScrollModeValue() != ScrollHorizontal {
					t.Errorf("got %v, want ScrollHorizontal", e.ScrollModeValue())
				}
			},
		},
		"ScrollMode both": {
			setup: func(e *Element) {
				WithScrollable(ScrollBoth)(e)
			},
			check: func(t *testing.T, e *Element) {
				if e.ScrollModeValue() != ScrollBoth {
					t.Errorf("got %v, want ScrollBoth", e.ScrollModeValue())
				}
			},
		},
		"not scrollable by default": {
			setup: func(e *Element) {},
			check: func(t *testing.T, e *Element) {
				if e.IsScrollable() {
					t.Error("expected IsScrollable() = false by default")
				}
			},
		},
		"ScrollTo with no content stays at zero": {
			setup: func(e *Element) {
				e.ScrollTo(5, 5)
			},
			check: func(t *testing.T, e *Element) {
				x, y := e.ScrollOffset()
				if x != 0 || y != 0 {
					t.Errorf("scroll offset = (%d, %d), want (0, 0) when no content", x, y)
				}
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			e := New(WithSize(20, 10))
			tt.setup(e)
			tt.check(t, e)
		})
	}
}

func TestElement_ScrollbarHidden(t *testing.T) {
	const width, height = 20, 5

	makeScrollable := func(opts ...Option) *Element {
		opts = append([]Option{
			WithSize(width, height),
			WithScrollable(ScrollVertical),
			WithDirection(Column),
		}, opts...)
		e := New(opts...)
		// More rows than viewport so a scrollbar would be needed.
		for i := 0; i < height*3; i++ {
			e.AddChild(New(WithHeight(1), WithBackground(NewStyle().Background(Red))))
		}
		return e
	}

	t.Run("default shows scrollbar and reserves gutter", func(t *testing.T) {
		e := makeScrollable()
		buf := NewBuffer(width, height)
		e.Render(buf, width, height)

		if !e.needsVerticalScrollbar() {
			t.Fatalf("expected scrollbar to be needed")
		}

		vw, _ := e.ViewportSize()
		if vw != width {
			t.Fatalf("viewport width = %d, want %d", vw, width)
		}

		gutter := buf.Cell(width-1, 0).Rune
		if gutter != '█' && gutter != '│' {
			t.Errorf("rightmost column rune = %q, want scrollbar glyph", gutter)
		}
	})

	t.Run("hidden skips scrollbar and reclaims gutter", func(t *testing.T) {
		e := makeScrollable(WithScrollbarHidden(true))
		buf := NewBuffer(width, height)
		e.Render(buf, width, height)

		if e.needsVerticalScrollbar() {
			t.Fatalf("expected scrollbar to be hidden")
		}

		for y := 0; y < height; y++ {
			r := buf.Cell(width-1, y).Rune
			if r == '█' || r == '│' {
				t.Errorf("row %d rightmost column = %q, expected child content (no scrollbar glyph)", y, r)
			}
		}
	})
}

func TestElement_ScrollToTop(t *testing.T) {
	e := New(
		WithSize(20, 5),
		WithScrollable(ScrollVertical),
		WithDirection(Column),
	)
	for i := 0; i < 20; i++ {
		e.AddChild(New(WithHeight(1)))
	}
	buf := NewBuffer(20, 5)
	e.Render(buf, 20, 5)

	e.ScrollTo(0, 10)
	e.ScrollToTop()
	_, y := e.ScrollOffset()
	if y != 0 {
		t.Errorf("after ScrollToTop, y = %d, want 0", y)
	}
}
