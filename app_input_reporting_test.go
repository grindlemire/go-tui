package tui

import "testing"

// Input reporting picks mouse reporting when mouse is enabled, and falls back to
// alternate-scroll in full-screen mode when mouse is off (so the wheel still
// scrolls while native selection and link clicking keep working). Inline mode
// with mouse off enables neither, since there is no alternate screen.
func TestInputReporting(t *testing.T) {
	type tc struct {
		mouseEnabled  bool
		inlineHeight  int
		wantMouse     bool
		wantAltScroll bool
	}

	tests := map[string]tc{
		"full-screen with mouse": {
			mouseEnabled:  true,
			inlineHeight:  0,
			wantMouse:     true,
			wantAltScroll: false,
		},
		"full-screen without mouse": {
			mouseEnabled:  false,
			inlineHeight:  0,
			wantMouse:     false,
			wantAltScroll: true,
		},
		"inline without mouse": {
			mouseEnabled:  false,
			inlineHeight:  5,
			wantMouse:     false,
			wantAltScroll: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			term := NewMockTerminal(80, 24)
			app := &App{
				terminal:     term,
				mouseEnabled: tt.mouseEnabled,
				inlineHeight: tt.inlineHeight,
			}

			app.enableInputReporting()
			if term.IsMouseEnabled() != tt.wantMouse {
				t.Errorf("after enable: mouse = %v, want %v", term.IsMouseEnabled(), tt.wantMouse)
			}
			if term.IsAltScrollEnabled() != tt.wantAltScroll {
				t.Errorf("after enable: altScroll = %v, want %v", term.IsAltScrollEnabled(), tt.wantAltScroll)
			}

			app.disableInputReporting()
			if term.IsMouseEnabled() {
				t.Error("after disable: mouse still enabled")
			}
			if term.IsAltScrollEnabled() {
				t.Error("after disable: altScroll still enabled")
			}
		})
	}
}
