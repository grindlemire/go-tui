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
	type tc struct {
		method func(Style) Style
		attr   Attr
	}

	tests := map[string]tc{
		"Bold":          {method: Style.Bold, attr: AttrBold},
		"Dim":           {method: Style.Dim, attr: AttrDim},
		"Italic":        {method: Style.Italic, attr: AttrItalic},
		"Underline":     {method: Style.Underline, attr: AttrUnderline},
		"Blink":         {method: Style.Blink, attr: AttrBlink},
		"Reverse":       {method: Style.Reverse, attr: AttrReverse},
		"Strikethrough": {method: Style.Strikethrough, attr: AttrStrikethrough},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			s := tt.method(NewStyle())
			if !s.HasAttr(tt.attr) {
				t.Errorf("%s() should set %v attribute", name, tt.attr)
			}
		})
	}
}

func TestStyle_Equal(t *testing.T) {
	type tc struct {
		a, b  Style
		equal bool
	}

	tests := map[string]tc{
		"empty styles": {
			a:     NewStyle(),
			b:     NewStyle(),
			equal: true,
		},
		"same foreground": {
			a:     NewStyle().Foreground(Red),
			b:     NewStyle().Foreground(Red),
			equal: true,
		},
		"different foreground": {
			a:     NewStyle().Foreground(Red),
			b:     NewStyle().Foreground(Blue),
			equal: false,
		},
		"same background": {
			a:     NewStyle().Background(Green),
			b:     NewStyle().Background(Green),
			equal: true,
		},
		"different background": {
			a:     NewStyle().Background(Green),
			b:     NewStyle().Background(Yellow),
			equal: false,
		},
		"same attributes": {
			a:     NewStyle().Bold().Italic(),
			b:     NewStyle().Bold().Italic(),
			equal: true,
		},
		"different attributes": {
			a:     NewStyle().Bold(),
			b:     NewStyle().Italic(),
			equal: false,
		},
		"full match": {
			a:     NewStyle().Foreground(Red).Background(Blue).Bold().Underline(),
			b:     NewStyle().Foreground(Red).Background(Blue).Bold().Underline(),
			equal: true,
		},
		"full mismatch on attr": {
			a:     NewStyle().Foreground(Red).Background(Blue).Bold(),
			b:     NewStyle().Foreground(Red).Background(Blue).Italic(),
			equal: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.a.Equal(tt.b); got != tt.equal {
				t.Errorf("Equal() = %v, want %v", got, tt.equal)
			}
			// Test symmetry
			if got := tt.b.Equal(tt.a); got != tt.equal {
				t.Errorf("(symmetric) Equal() = %v, want %v", got, tt.equal)
			}
		})
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
	type tc struct {
		attr Attr
	}

	tests := map[string]tc{
		"Bold":          {attr: AttrBold},
		"Dim":           {attr: AttrDim},
		"Italic":        {attr: AttrItalic},
		"Underline":     {attr: AttrUnderline},
		"Blink":         {attr: AttrBlink},
		"Reverse":       {attr: AttrReverse},
		"Strikethrough": {attr: AttrStrikethrough},
	}

	// Collect all attrs for overlap and combination tests
	var allAttrs []Attr
	var attrNames []string
	for name, tt := range tests {
		allAttrs = append(allAttrs, tt.attr)
		attrNames = append(attrNames, name)
	}

	// Verify attributes are distinct bit flags
	for i, a := range allAttrs {
		for j, b := range allAttrs {
			if i != j && a&b != 0 {
				t.Errorf("Attr %s and %s overlap in bits", attrNames[i], attrNames[j])
			}
		}
	}

	// Verify we can combine all attributes
	var combined Attr
	for _, a := range allAttrs {
		combined |= a
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if combined&tt.attr == 0 {
				t.Errorf("Combined attrs missing %v", tt.attr)
			}
		})
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
