package tui

import "testing"

func TestWithOverlay_SetsField(t *testing.T) {
	e := New(WithOverlay(true))
	if !e.IsOverlay() {
		t.Error("expected overlay to be true")
	}

	e2 := New(WithOverlay(false))
	if e2.IsOverlay() {
		t.Error("expected overlay to be false")
	}
}

func TestLayoutChildren_ExcludesOverlay(t *testing.T) {
	parent := New()
	normal := New(WithText("visible"))
	overlay := New(WithOverlay(true), WithText("overlay"))
	parent.AddChild(normal)
	parent.AddChild(overlay)

	children := parent.LayoutChildren()
	if len(children) != 1 {
		t.Fatalf("expected 1 layout child, got %d", len(children))
	}
}

func TestRenderElement_SkipsOverlayChildren(t *testing.T) {
	parent := New(WithWidth(20), WithHeight(3))
	normal := New(WithText("visible"))
	overlay := New(WithOverlay(true), WithText("hidden"))
	parent.AddChild(normal)
	parent.AddChild(overlay)

	buf := NewBuffer(20, 3)
	parent.Render(buf, 20, 3)

	content := buf.String()
	if !containsSubstring(content, "visible") {
		t.Error("expected normal child to be rendered")
	}
	if containsSubstring(content, "hidden") {
		t.Error("expected overlay child to NOT be rendered")
	}
}

func TestApply_SetsOptions(t *testing.T) {
	e := New()
	e.Apply(WithOverlay(true), WithHidden(true))
	if !e.IsOverlay() {
		t.Error("expected overlay after Apply")
	}
	if !e.Hidden() {
		t.Error("expected hidden after Apply")
	}
}
