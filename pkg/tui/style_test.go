package tui

import (
	"testing"
)

func TestNewStyle(t *testing.T) {
	s := NewStyle()

	// Should have default colors
	if !s.Fg.IsDefault() {
		t.Error("NewStyle().Fg should be default color")
	}
	if !s.Bg.IsDefault() {
		t.Error("NewStyle().Bg should be default color")
	}

	// Should have no attributes
	if s.Attrs != AttrNone {
		t.Errorf("NewStyle().Attrs = %v, want AttrNone", s.Attrs)
	}
}

func TestStyle_Foreground(t *testing.T) {
	s := NewStyle().Foreground(Red)

	if !s.Fg.Equal(Red) {
		t.Errorf("Foreground(Red).Fg = %v, want Red", s.Fg)
	}
	if !s.Bg.IsDefault() {
		t.Error("Foreground() should not affect background")
	}
}

func TestStyle_Background(t *testing.T) {
	s := NewStyle().Background(Blue)

	if !s.Bg.Equal(Blue) {
		t.Errorf("Background(Blue).Bg = %v, want Blue", s.Bg)
	}
	if !s.Fg.IsDefault() {
		t.Error("Background() should not affect foreground")
	}
}

func TestStyle_FluentChaining(t *testing.T) {
	s := NewStyle().
		Foreground(Red).
		Background(Blue).
		Bold().
		Italic().
		Underline()

	if !s.Fg.Equal(Red) {
		t.Errorf("chained style Fg = %v, want Red", s.Fg)
	}
	if !s.Bg.Equal(Blue) {
		t.Errorf("chained style Bg = %v, want Blue", s.Bg)
	}
	if !s.HasAttr(AttrBold) {
		t.Error("chained style should have AttrBold")
	}
	if !s.HasAttr(AttrItalic) {
		t.Error("chained style should have AttrItalic")
	}
	if !s.HasAttr(AttrUnderline) {
		t.Error("chained style should have AttrUnderline")
	}
}

func TestStyle_AllAttributes(t *testing.T) {
	attrs := []struct {
		name   string
		method func(Style) Style
		attr   Attr
	}{
		{"Bold", Style.Bold, AttrBold},
		{"Dim", Style.Dim, AttrDim},
		{"Italic", Style.Italic, AttrItalic},
		{"Underline", Style.Underline, AttrUnderline},
		{"Blink", Style.Blink, AttrBlink},
		{"Reverse", Style.Reverse, AttrReverse},
		{"Strikethrough", Style.Strikethrough, AttrStrikethrough},
	}

	for _, tt := range attrs {
		s := tt.method(NewStyle())
		if !s.HasAttr(tt.attr) {
			t.Errorf("%s() should set %v attribute", tt.name, tt.attr)
		}
	}
}

func TestStyle_Equal(t *testing.T) {
	tests := []struct {
		name  string
		a, b  Style
		equal bool
	}{
		{
			"empty styles",
			NewStyle(),
			NewStyle(),
			true,
		},
		{
			"same foreground",
			NewStyle().Foreground(Red),
			NewStyle().Foreground(Red),
			true,
		},
		{
			"different foreground",
			NewStyle().Foreground(Red),
			NewStyle().Foreground(Blue),
			false,
		},
		{
			"same background",
			NewStyle().Background(Green),
			NewStyle().Background(Green),
			true,
		},
		{
			"different background",
			NewStyle().Background(Green),
			NewStyle().Background(Yellow),
			false,
		},
		{
			"same attributes",
			NewStyle().Bold().Italic(),
			NewStyle().Bold().Italic(),
			true,
		},
		{
			"different attributes",
			NewStyle().Bold(),
			NewStyle().Italic(),
			false,
		},
		{
			"full match",
			NewStyle().Foreground(Red).Background(Blue).Bold().Underline(),
			NewStyle().Foreground(Red).Background(Blue).Bold().Underline(),
			true,
		},
		{
			"full mismatch on attr",
			NewStyle().Foreground(Red).Background(Blue).Bold(),
			NewStyle().Foreground(Red).Background(Blue).Italic(),
			false,
		},
	}

	for _, tt := range tests {
		if got := tt.a.Equal(tt.b); got != tt.equal {
			t.Errorf("%s: Equal() = %v, want %v", tt.name, got, tt.equal)
		}
		// Test symmetry
		if got := tt.b.Equal(tt.a); got != tt.equal {
			t.Errorf("%s (symmetric): Equal() = %v, want %v", tt.name, got, tt.equal)
		}
	}
}

func TestStyle_HasAttr(t *testing.T) {
	s := NewStyle().Bold().Italic()

	// Should have individual attributes
	if !s.HasAttr(AttrBold) {
		t.Error("HasAttr(AttrBold) should return true")
	}
	if !s.HasAttr(AttrItalic) {
		t.Error("HasAttr(AttrItalic) should return true")
	}

	// Should have combined attributes
	if !s.HasAttr(AttrBold | AttrItalic) {
		t.Error("HasAttr(AttrBold|AttrItalic) should return true")
	}

	// Should not have attributes not set
	if s.HasAttr(AttrUnderline) {
		t.Error("HasAttr(AttrUnderline) should return false")
	}

	// AttrNone should always return true (empty mask)
	if !s.HasAttr(AttrNone) {
		t.Error("HasAttr(AttrNone) should return true")
	}
}

func TestStyle_Immutability(t *testing.T) {
	original := NewStyle()
	modified := original.Bold().Foreground(Red)

	// Original should be unchanged
	if original.HasAttr(AttrBold) {
		t.Error("original style should not be modified")
	}
	if !original.Fg.IsDefault() {
		t.Error("original style foreground should be unchanged")
	}

	// Modified should have changes
	if !modified.HasAttr(AttrBold) {
		t.Error("modified style should have bold")
	}
	if !modified.Fg.Equal(Red) {
		t.Error("modified style should have red foreground")
	}
}

func TestAttr_BitfieldValues(t *testing.T) {
	// Verify attributes are distinct bit flags
	attrs := []Attr{AttrBold, AttrDim, AttrItalic, AttrUnderline, AttrBlink, AttrReverse, AttrStrikethrough}

	for i, a := range attrs {
		for j, b := range attrs {
			if i != j && a&b != 0 {
				t.Errorf("Attr %d and %d overlap in bits", i, j)
			}
		}
	}

	// Verify we can combine all attributes
	var combined Attr
	for _, a := range attrs {
		combined |= a
	}

	for _, a := range attrs {
		if combined&a == 0 {
			t.Errorf("Combined attrs missing %v", a)
		}
	}
}

func TestStyle_ZeroValue(t *testing.T) {
	var s Style

	// Zero value should be equivalent to NewStyle()
	if !s.Equal(NewStyle()) {
		t.Error("zero value Style should equal NewStyle()")
	}

	// Zero value should have default colors
	if !s.Fg.IsDefault() {
		t.Error("zero value Style.Fg should be default")
	}
	if !s.Bg.IsDefault() {
		t.Error("zero value Style.Bg should be default")
	}

	// Zero value should have no attributes
	if s.Attrs != AttrNone {
		t.Errorf("zero value Style.Attrs = %v, want AttrNone", s.Attrs)
	}
}
