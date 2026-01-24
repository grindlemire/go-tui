package element

import (
	"testing"

	"github.com/grindlemire/go-tui/pkg/layout"
	"github.com/grindlemire/go-tui/pkg/tui"
)

func TestNewText_CreatesWithContent(t *testing.T) {
	text := NewText("Hello, World!")

	if text.Content() != "Hello, World!" {
		t.Errorf("NewText content = %q, want %q", text.Content(), "Hello, World!")
	}

	// Should have embedded Element
	if text.Element == nil {
		t.Fatal("NewText should have embedded Element")
	}

	// Default alignment should be left
	if text.Align() != TextAlignLeft {
		t.Errorf("NewText align = %v, want TextAlignLeft", text.Align())
	}
}

func TestNewText_WithOptions(t *testing.T) {
	type tc struct {
		opts    []TextOption
		check   func(*Text) bool
		message string
	}

	tests := map[string]tc{
		"WithTextStyle": {
			opts: []TextOption{WithTextStyle(tui.NewStyle().Foreground(tui.Red))},
			check: func(text *Text) bool {
				return text.ContentStyle().Fg == tui.Red
			},
			message: "WithTextStyle should set content style",
		},
		"WithTextAlign center": {
			opts:    []TextOption{WithTextAlign(TextAlignCenter)},
			check:   func(text *Text) bool { return text.Align() == TextAlignCenter },
			message: "WithTextAlign should set alignment",
		},
		"WithTextAlign right": {
			opts:    []TextOption{WithTextAlign(TextAlignRight)},
			check:   func(text *Text) bool { return text.Align() == TextAlignRight },
			message: "WithTextAlign should set alignment",
		},
		"WithElementOption": {
			opts: []TextOption{WithElementOption(WithWidth(50))},
			check: func(text *Text) bool {
				return text.Element.style.Width == layout.Fixed(50)
			},
			message: "WithElementOption should apply to embedded Element",
		},
		"multiple options": {
			opts: []TextOption{
				WithTextStyle(tui.NewStyle().Bold()),
				WithTextAlign(TextAlignCenter),
				WithElementOption(WithBorder(tui.BorderSingle)),
			},
			check: func(text *Text) bool {
				return text.ContentStyle().HasAttr(tui.AttrBold) &&
					text.Align() == TextAlignCenter &&
					text.Element.border == tui.BorderSingle
			},
			message: "multiple options should all be applied",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			text := NewText("test", tt.opts...)
			if !tt.check(text) {
				t.Error(tt.message)
			}
		})
	}
}

func TestText_SetContent(t *testing.T) {
	text := NewText("initial")

	text.SetContent("updated")

	if text.Content() != "updated" {
		t.Errorf("SetContent: Content() = %q, want %q", text.Content(), "updated")
	}
}

func TestText_SetContentStyle(t *testing.T) {
	text := NewText("test")

	style := tui.NewStyle().Foreground(tui.Green).Bold()
	text.SetContentStyle(style)

	if text.ContentStyle().Fg != tui.Green {
		t.Error("SetContentStyle should update foreground color")
	}
	if !text.ContentStyle().HasAttr(tui.AttrBold) {
		t.Error("SetContentStyle should update bold attribute")
	}
}

func TestText_SetAlign(t *testing.T) {
	text := NewText("test")

	text.SetAlign(TextAlignRight)

	if text.Align() != TextAlignRight {
		t.Errorf("SetAlign: Align() = %v, want TextAlignRight", text.Align())
	}
}

func TestText_InheritsElementBehavior(t *testing.T) {
	// Text should inherit all Element behavior through embedding
	text := NewText("test", WithElementOption(WithSize(100, 50)))

	text.Calculate(200, 200)

	rect := text.Rect()
	if rect.Width != 100 || rect.Height != 50 {
		t.Errorf("Text.Rect() = %dx%d, want 100x50", rect.Width, rect.Height)
	}

	// Should have dirty behavior
	text.dirty = false
	text.MarkDirty()
	if !text.IsDirty() {
		t.Error("Text should support MarkDirty through embedding")
	}
}

func TestText_CanBeAddedAsChild(t *testing.T) {
	parent := New(WithSize(100, 50))
	text := NewText("child text")

	// Text's Element can be added as a child
	parent.AddChild(text.Element)

	if len(parent.Children()) != 1 {
		t.Errorf("parent should have 1 child, got %d", len(parent.Children()))
	}

	if text.Element.Parent() != parent {
		t.Error("text.Element.Parent() should be parent")
	}
}
