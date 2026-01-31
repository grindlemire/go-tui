package lsp

import (
	"testing"
)

func TestGetLineText(t *testing.T) {
	type tc struct {
		content string
		line    int
		want    string
	}

	tests := map[string]tc{
		"first line": {
			content: "package test\n\ntempl Foo() {\n}\n",
			line:    0,
			want:    "package test",
		},
		"empty line": {
			content: "package test\n\ntempl Foo() {\n}\n",
			line:    1,
			want:    "",
		},
		"third line": {
			content: "package test\n\ntempl Foo() {\n}\n",
			line:    2,
			want:    "templ Foo() {",
		},
		"last line no newline": {
			content: "line1\nline2",
			line:    1,
			want:    "line2",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := getLineText(tt.content, tt.line)
			if got != tt.want {
				t.Errorf("getLineText line %d = %q, want %q", tt.line, got, tt.want)
			}
		})
	}
}

func TestGetWordAtOffset(t *testing.T) {
	type tc struct {
		content string
		offset  int
		want    string
	}

	tests := map[string]tc{
		"simple word": {
			content: "hello world",
			offset:  2,
			want:    "hello",
		},
		"second word": {
			content: "hello world",
			offset:  7,
			want:    "world",
		},
		"hyphenated": {
			content: "class=\"flex-col\"",
			offset:  10,
			want:    "flex-col",
		},
		"at symbol": {
			content: "@for i := range items",
			offset:  2,
			want:    "@for",
		},
		"hash prefix": {
			content: "<div #Header>",
			offset:  7,
			want:    "#Header",
		},
		"out of bounds": {
			content: "test",
			offset:  -1,
			want:    "",
		},
		"past end": {
			content: "test",
			offset:  10,
			want:    "",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := getWordAtOffset(tt.content, tt.offset)
			if got != tt.want {
				t.Errorf("getWordAtOffset(%q, %d) = %q, want %q", tt.content, tt.offset, got, tt.want)
			}
		})
	}
}

func TestIsOffsetInGoExpr(t *testing.T) {
	type tc struct {
		content string
		offset  int
		want    bool
	}

	tests := map[string]tc{
		"inside braces": {
			content: "<span>{count}</span>",
			offset:  8,
			want:    true,
		},
		"outside braces": {
			content: "<span>{count}</span>",
			offset:  3,
			want:    false,
		},
		"at start": {
			content: "{expr}",
			offset:  0,
			want:    false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := isOffsetInGoExpr(tt.content, tt.offset)
			if got != tt.want {
				t.Errorf("isOffsetInGoExpr(%q, %d) = %v, want %v", tt.content, tt.offset, got, tt.want)
			}
		})
	}
}

func TestIsOffsetInClassAttr(t *testing.T) {
	type tc struct {
		content string
		offset  int
		want    bool
	}

	tests := map[string]tc{
		"inside class value": {
			content: `<div class="flex-col">`,
			offset:  15,
			want:    true,
		},
		"outside class": {
			content: `<div class="flex-col">`,
			offset:  3,
			want:    false,
		},
		"after closing quote": {
			content: `<div class="flex-col" id="x">`,
			offset:  25,
			want:    false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := isOffsetInClassAttr(tt.content, tt.offset)
			if got != tt.want {
				t.Errorf("isOffsetInClassAttr(%q, %d) = %v, want %v", tt.content, tt.offset, got, tt.want)
			}
		})
	}
}

func TestIsOffsetInElementTag(t *testing.T) {
	type tc struct {
		content string
		offset  int
		want    bool
	}

	tests := map[string]tc{
		"inside tag": {
			content: "<div class=\"p-1\">",
			offset:  6,
			want:    true,
		},
		"outside tag": {
			content: "<div>text</div>",
			offset:  8,
			want:    false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := isOffsetInElementTag(tt.content, tt.offset)
			if got != tt.want {
				t.Errorf("isOffsetInElementTag(%q, %d) = %v, want %v", tt.content, tt.offset, got, tt.want)
			}
		})
	}
}
