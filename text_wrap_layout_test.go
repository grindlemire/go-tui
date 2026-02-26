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
