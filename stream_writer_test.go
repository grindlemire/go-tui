package tui

import (
	"strings"
	"testing"
)

func TestStreamWriter_WriteStyled(t *testing.T) {
	type tc struct {
		text     string
		style    Style
		caps     Capabilities
		contains string // expected substring in output
	}

	tests := map[string]tc{
		"bold red text": {
			text:     "hello",
			style:    NewStyle().Bold().Foreground(Red),
			caps:     Capabilities{Colors: Color16, TrueColor: false},
			contains: "\x1b[0;1;31mhello\x1b[0m",
		},
		"dim italic": {
			text:     "dim",
			style:    NewStyle().Dim().Italic(),
			caps:     Capabilities{Colors: Color16},
			contains: "\x1b[0;2;3mdim\x1b[0m",
		},
		"rgb foreground with truecolor": {
			text:     "x",
			style:    NewStyle().Foreground(RGBColor(255, 128, 0)),
			caps:     Capabilities{Colors: ColorTrue, TrueColor: true},
			contains: "\x1b[0;38;2;255;128;0mx\x1b[0m",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var buf strings.Builder
			sw := &StreamWriter{
				w:    &writerCloserAdapter{w: &buf},
				caps: tt.caps,
			}

			n, err := sw.WriteStyled(tt.text, tt.style)
			if err != nil {
				t.Fatalf("WriteStyled error: %v", err)
			}
			if n != buf.Len() {
				t.Errorf("WriteStyled returned n=%d, buf has %d bytes", n, buf.Len())
			}
			if !strings.Contains(buf.String(), tt.contains) {
				t.Errorf("output %q does not contain %q", buf.String(), tt.contains)
			}
		})
	}
}

func TestStreamWriter_WriteGradient(t *testing.T) {
	type tc struct {
		text  string
		width int
	}

	tests := map[string]tc{
		"single char": {
			text:  "A",
			width: 80,
		},
		"multiple chars": {
			text:  "ABC",
			width: 80,
		},
		"with newline": {
			text:  "AB\nCD",
			width: 80,
		},
	}

	caps := Capabilities{Colors: ColorTrue, TrueColor: true}
	grad := NewGradient(Red, Blue)

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var buf strings.Builder
			sw := &StreamWriter{
				w:     &writerCloserAdapter{w: &buf},
				width: tt.width,
				caps:  caps,
			}

			n, err := sw.WriteGradient(tt.text, grad)
			if err != nil {
				t.Fatalf("WriteGradient error: %v", err)
			}
			if n != buf.Len() {
				t.Errorf("WriteGradient returned n=%d, buf has %d bytes", n, buf.Len())
			}

			out := buf.String()
			// Each non-newline character should produce an SGR sequence.
			nonNewline := 0
			for _, r := range tt.text {
				if r != '\n' {
					nonNewline++
				}
			}
			// Count SGR sequences (ESC[...m patterns).
			sgrCount := strings.Count(out, "\x1b[0")
			// Expect one per character + one trailing reset.
			if sgrCount < nonNewline {
				t.Errorf("expected at least %d SGR sequences, got %d in %q", nonNewline, sgrCount, out)
			}

			// Verify newlines pass through.
			if strings.Count(tt.text, "\n") != strings.Count(out, "\n") {
				t.Errorf("newline count mismatch: text has %d, output has %d",
					strings.Count(tt.text, "\n"), strings.Count(out, "\n"))
			}
		})
	}
}

func TestStreamWriter_WriteGradientWithBase(t *testing.T) {
	var buf strings.Builder
	caps := Capabilities{Colors: ColorTrue, TrueColor: true}
	sw := &StreamWriter{
		w:     &writerCloserAdapter{w: &buf},
		width: 80,
		caps:  caps,
	}

	base := NewStyle().Bold().Background(Blue)
	grad := NewGradient(Red, Green)

	_, err := sw.WriteGradient("AB", grad, base)
	if err != nil {
		t.Fatalf("WriteGradient error: %v", err)
	}

	out := buf.String()
	// Should contain bold attribute (;1).
	if !strings.Contains(out, ";1") {
		t.Errorf("expected bold attribute in output %q", out)
	}
	// Should contain background color for blue (;44 for basic ANSI blue).
	// With truecolor caps, ANSI blue (index 4) uses basic code 44.
	if !strings.Contains(out, ";44") {
		t.Errorf("expected background color 44 in output %q", out)
	}
}

func TestStreamWriter_NopWriter(t *testing.T) {
	sw := &StreamWriter{
		w:   &nopStreamWriter{},
		nop: true,
	}

	n, err := sw.WriteStyled("hello", NewStyle().Bold())
	if err != nil {
		t.Fatalf("WriteStyled error: %v", err)
	}
	if n != 5 {
		t.Errorf("WriteStyled returned n=%d, want 5", n)
	}

	n, err = sw.WriteGradient("world", NewGradient(Red, Blue))
	if err != nil {
		t.Fatalf("WriteGradient error: %v", err)
	}
	if n != 5 {
		t.Errorf("WriteGradient returned n=%d, want 5", n)
	}

	n, err = sw.Write([]byte("raw"))
	if err != nil {
		t.Fatalf("Write error: %v", err)
	}
	if n != 3 {
		t.Errorf("Write returned n=%d, want 3", n)
	}

	if err := sw.Close(); err != nil {
		t.Fatalf("Close error: %v", err)
	}
}

func TestStreamWriter_ColumnTracking(t *testing.T) {
	type tc struct {
		writes   []string
		width    int
		wantCol  int
	}

	tests := map[string]tc{
		"simple advance": {
			writes:  []string{"ABC"},
			width:   80,
			wantCol: 3,
		},
		"newline resets": {
			writes:  []string{"AB\nC"},
			width:   80,
			wantCol: 1,
		},
		"wraps at width": {
			writes:  []string{"ABCDE"},
			width:   5,
			wantCol: 0,
		},
		"multi write accumulates": {
			writes:  []string{"AB", "CD"},
			width:   80,
			wantCol: 4,
		},
	}

	caps := Capabilities{Colors: ColorTrue, TrueColor: true}
	grad := NewGradient(Red, Blue)

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var buf strings.Builder
			sw := &StreamWriter{
				w:     &writerCloserAdapter{w: &buf},
				width: tt.width,
				caps:  caps,
			}

			for _, s := range tt.writes {
				sw.WriteGradient(s, grad)
			}

			if sw.col != tt.wantCol {
				t.Errorf("col = %d, want %d", sw.col, tt.wantCol)
			}
		})
	}
}

func TestStreamWriter_BackwardCompat(t *testing.T) {
	app, _ := newInlineTestApp(80, 24, 3)
	sw := app.StreamAbove()

	// Plain Write still works.
	n, err := sw.Write([]byte("hello"))
	if err != nil {
		t.Fatalf("Write error: %v", err)
	}
	if n != 5 {
		t.Errorf("Write returned n=%d, want 5", n)
	}

	// Close still works.
	if err := sw.Close(); err != nil {
		t.Fatalf("Close error: %v", err)
	}
}

// writerCloserAdapter wraps a strings.Builder to satisfy io.WriteCloser.
type writerCloserAdapter struct {
	w *strings.Builder
}

func (a *writerCloserAdapter) Write(p []byte) (int, error) {
	return a.w.Write(p)
}

func (a *writerCloserAdapter) Close() error {
	return nil
}
