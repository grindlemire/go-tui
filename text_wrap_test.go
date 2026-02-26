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
