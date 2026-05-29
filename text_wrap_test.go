package tui

import (
	"testing"
)

func TestWrapText(t *testing.T) {
	type tc struct {
		text     string
		maxWidth int
		want     []string
	}

	tests := map[string]tc{
		"empty string": {
			text:     "",
			maxWidth: 10,
			want:     []string{""},
		},
		"fits on one line": {
			text:     "hello",
			maxWidth: 10,
			want:     []string{"hello"},
		},
		"wraps at word boundary": {
			text:     "hello world",
			maxWidth: 7,
			want:     []string{"hello", "world"},
		},
		"multiple wraps": {
			text:     "the quick brown fox",
			maxWidth: 10,
			want:     []string{"the quick", "brown fox"},
		},
		"long word breaks mid-word": {
			text:     "abcdefghij",
			maxWidth: 5,
			want:     []string{"abcde", "fghij"},
		},
		"long word after short word": {
			text:     "hi abcdefghij",
			maxWidth: 5,
			want:     []string{"hi", "abcde", "fghij"},
		},
		"preserves newlines": {
			text:     "line1\nline2",
			maxWidth: 20,
			want:     []string{"line1", "line2"},
		},
		"wraps within newline sections": {
			text:     "hello world\nfoo bar",
			maxWidth: 7,
			want:     []string{"hello", "world", "foo bar"},
		},
		"zero width": {
			text:     "hello",
			maxWidth: 0,
			want:     []string{""},
		},
		"width of 1": {
			text:     "hi",
			maxWidth: 1,
			want:     []string{"h", "i"},
		},
		"exact fit": {
			text:     "hello",
			maxWidth: 5,
			want:     []string{"hello"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := wrapText(tt.text, tt.maxWidth)
			if len(got) != len(tt.want) {
				t.Fatalf("wrapText(%q, %d) = %v (len %d), want %v (len %d)",
					tt.text, tt.maxWidth, got, len(got), tt.want, len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("line %d: got %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestWrapSpans_KeepsStyleAcrossWrap(t *testing.T) {
	bold := NewStyle().Bold()
	spans := []TextSpan{
		{Text: "aa "},
		{Text: "bbbb cccc", Style: bold}, // two bold words
		{Text: " dd"},
	}
	// Width 6 forces a break inside the bold run.
	lines := wrapSpans(spans, 6)
	if len(lines) < 2 {
		t.Fatalf("expected multiple lines, got %d: %+v", len(lines), lines)
	}
	// Every segment whose text is a bold word must carry bold on every line.
	for li, line := range lines {
		for _, seg := range line {
			if seg.Text == "bbbb" || seg.Text == "cccc" {
				if seg.Style.Attrs&AttrBold == 0 {
					t.Errorf("line %d: %q lost bold", li, seg.Text)
				}
			}
		}
	}
}

func TestWrapSpans_MergesAdjacentSameStyle(t *testing.T) {
	// Two plain spans with words that fit on one line should merge.
	spans := []TextSpan{{Text: "foo "}, {Text: "bar"}}
	lines := wrapSpans(spans, 40)
	if len(lines) != 1 {
		t.Fatalf("want 1 line, got %d", len(lines))
	}
	if len(lines[0]) != 1 {
		t.Errorf("adjacent same-style segments should merge into 1, got %d: %+v", len(lines[0]), lines[0])
	}
	if lines[0][0].Text != "foo bar" {
		t.Errorf("merged text = %q, want \"foo bar\"", lines[0][0].Text)
	}
}

func TestWrapSpans_Empty(t *testing.T) {
	if got := wrapSpans(nil, 10); len(got) != 1 || len(got[0]) != 0 {
		t.Errorf("empty spans should give one empty line, got %+v", got)
	}
}
