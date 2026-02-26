package tui

import (
	"testing"
)

func TestIntegration_TextWrapLayout(t *testing.T) {
	type tc struct {
		text        string
		parentWidth int
		wantHeight  int
	}

	tests := map[string]tc{
		"text wraps increases element height": {
			text:        "hello world foo",
			parentWidth: 7,
			wantHeight:  3, // "hello", "world", "foo"
		},
		"no wrap when text fits": {
			text:        "hello",
			parentWidth: 20,
			wantHeight:  1,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			child := New(WithText(tt.text))
			parent := New(
				WithWidth(tt.parentWidth),
				WithHeight(20),
				WithDirection(Column),
			)
			parent.AddChild(child)
			parent.Calculate(tt.parentWidth, 20)

			childRect := child.Rect()
			if childRect.Height != tt.wantHeight {
				t.Errorf("child height = %d, want %d (rect=%+v)", childRect.Height, tt.wantHeight, childRect)
			}
		})
	}
}

// TestIntegration_TextWrapLayout_NestedContainer tests that a container with
// a wrapping text child pushes subsequent siblings down correctly.
// This reproduces the bug where the bordered "Text Elements" section in example 05
// didn't expand to fit wrapped text, pushing the <hr> out of the container.
func TestIntegration_TextWrapLayout_NestedContainer(t *testing.T) {
	// Simulate the structure from example 05:
	// <div class="flex-col" width=30 height=50>     (outer)
	//   <div class="flex-col border-rounded p-1">   (section)
	//     <span>Title</span>
	//     <p>Long paragraph text that wraps...</p>
	//     <hr />
	//     <span>After HR</span>
	//   </div>
	//   <span>Below section</span>
	// </div>
	longText := "Paragraph text wraps automatically when the content exceeds the available width"

	title := New(WithText("Title"))
	paragraph := New(WithText(longText))
	hr := New(WithText("──────────")) // simulate HR
	afterHR := New(WithText("After HR"))

	section := New(
		WithDirection(Column),
		WithBorder(BorderRounded),
		WithPadding(1),
	)
	section.AddChild(title, paragraph, hr, afterHR)

	belowSection := New(WithText("Below section"))

	outer := New(
		WithWidth(30),
		WithHeight(50),
		WithDirection(Column),
	)
	outer.AddChild(section, belowSection)
	outer.Calculate(30, 50)

	// The paragraph wraps within 30 - 2 border - 2 padding = 26 chars content width
	wrappedLines := wrapText(longText, 26)

	// Section height should include: 1 (title) + len(wrappedLines) (paragraph) +
	// 1 (hr) + 1 (afterHR) + 2 (padding) + 2 (border)
	expectedSectionHeight := len(wrappedLines) + 1 + 1 + 1 + 2 + 2
	sectionRect := section.Rect()
	if sectionRect.Height != expectedSectionHeight {
		t.Errorf("section height = %d, want %d (wrappedLines=%d, text=%q)",
			sectionRect.Height, expectedSectionHeight, len(wrappedLines), longText)
	}

	// "Below section" should start right after the section
	belowRect := belowSection.Rect()
	if belowRect.Y != sectionRect.Bottom() {
		t.Errorf("belowSection.Y = %d, want %d (section bottom)", belowRect.Y, sectionRect.Bottom())
	}

	// The HR should be visible within the section (not pushed out)
	hrRect := hr.Rect()
	sectionContentBottom := section.ContentRect().Bottom()
	if hrRect.Bottom() > sectionContentBottom {
		t.Errorf("HR bottom (%d) exceeds section content bottom (%d) — HR pushed out of container",
			hrRect.Bottom(), sectionContentBottom)
	}
}

func TestIntegration_TextWrapLayout_RowDirection(t *testing.T) {
	type tc struct {
		text       string
		childWidth int
		wantHeight int
	}

	tests := map[string]tc{
		"row child wraps and gets taller": {
			text:       "hello world foo",
			childWidth: 7,
			wantHeight: 3, // "hello", "world", "foo"
		},
		"row child no wrap when fits": {
			text:       "hello",
			childWidth: 20,
			wantHeight: 1,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			child := New(WithText(tt.text), WithWidth(tt.childWidth))
			parent := New(
				WithSize(40, 20),
				WithDirection(Row),
				WithAlign(AlignStart), // Don't stretch, so child height reflects content
			)
			parent.AddChild(child)
			parent.Calculate(40, 20)

			childRect := child.Rect()
			if childRect.Height != tt.wantHeight {
				t.Errorf("child height = %d, want %d (rect=%+v)", childRect.Height, tt.wantHeight, childRect)
			}
		})
	}
}
