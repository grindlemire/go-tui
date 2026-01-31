package tui

import (
	"os"
	"testing"
)

// testEnvHelper saves and restores environment variables for testing.
type testEnvHelper struct {
	saved map[string]string
}

func newTestEnvHelper() *testEnvHelper {
	return &testEnvHelper{saved: make(map[string]string)}
}

func (h *testEnvHelper) Set(key, value string) {
	if _, exists := h.saved[key]; !exists {
		h.saved[key] = os.Getenv(key)
	}
	os.Setenv(key, value)
}

func (h *testEnvHelper) Clear(key string) {
	if _, exists := h.saved[key]; !exists {
		h.saved[key] = os.Getenv(key)
	}
	os.Unsetenv(key)
}

func (h *testEnvHelper) Restore() {
	for key, value := range h.saved {
		if value == "" {
			os.Unsetenv(key)
		} else {
			os.Setenv(key, value)
		}
	}
}

func clearTermEnvVars(env *testEnvHelper) {
	env.Clear("TERM")
	env.Clear("COLORTERM")
	env.Clear("WT_SESSION")
	env.Clear("ITERM_SESSION_ID")
	env.Clear("KITTY_WINDOW_ID")
	env.Clear("KONSOLE_VERSION")
	env.Clear("VTE_VERSION")
}

func TestDetectCapabilities_TERM256Color(t *testing.T) {
	env := newTestEnvHelper()
	defer env.Restore()
	clearTermEnvVars(env)

	env.Set("TERM", "xterm-256color")
	caps := DetectCapabilities()

	if caps.Colors != Color256 {
		t.Errorf("TERM=xterm-256color: Colors = %v, want Color256", caps.Colors)
	}
	if caps.TrueColor {
		t.Error("TERM=xterm-256color should not enable TrueColor")
	}
	if !caps.Unicode {
		t.Error("Unicode should be enabled by default")
	}
	if !caps.AltScreen {
		t.Error("AltScreen should be enabled by default")
	}
}

func TestDetectCapabilities_COLORTERMTruecolor(t *testing.T) {
	env := newTestEnvHelper()
	defer env.Restore()
	clearTermEnvVars(env)

	env.Set("TERM", "xterm-256color")
	env.Set("COLORTERM", "truecolor")
	caps := DetectCapabilities()

	if caps.Colors != ColorTrue {
		t.Errorf("COLORTERM=truecolor: Colors = %v, want ColorTrue", caps.Colors)
	}
	if !caps.TrueColor {
		t.Error("COLORTERM=truecolor should enable TrueColor")
	}
}

func TestDetectCapabilities_COLORTERM24bit(t *testing.T) {
	env := newTestEnvHelper()
	defer env.Restore()
	clearTermEnvVars(env)

	env.Set("COLORTERM", "24bit")
	caps := DetectCapabilities()

	if caps.Colors != ColorTrue {
		t.Errorf("COLORTERM=24bit: Colors = %v, want ColorTrue", caps.Colors)
	}
	if !caps.TrueColor {
		t.Error("COLORTERM=24bit should enable TrueColor")
	}
}

func TestDetectCapabilities_TERMDumb(t *testing.T) {
	env := newTestEnvHelper()
	defer env.Restore()
	clearTermEnvVars(env)

	env.Set("TERM", "dumb")
	caps := DetectCapabilities()

	if caps.Colors != ColorNone {
		t.Errorf("TERM=dumb: Colors = %v, want ColorNone", caps.Colors)
	}
	if caps.TrueColor {
		t.Error("TERM=dumb should not enable TrueColor")
	}
	if caps.Unicode {
		t.Error("TERM=dumb should disable Unicode")
	}
	if caps.AltScreen {
		t.Error("TERM=dumb should disable AltScreen")
	}
}

func TestDetectCapabilities_WTSession(t *testing.T) {
	env := newTestEnvHelper()
	defer env.Restore()
	clearTermEnvVars(env)

	env.Set("WT_SESSION", "some-session-id")
	caps := DetectCapabilities()

	if caps.Colors != ColorTrue {
		t.Errorf("WT_SESSION set: Colors = %v, want ColorTrue", caps.Colors)
	}
	if !caps.TrueColor {
		t.Error("WT_SESSION should enable TrueColor")
	}
}

func TestDetectCapabilities_ITerm(t *testing.T) {
	env := newTestEnvHelper()
	defer env.Restore()
	clearTermEnvVars(env)

	env.Set("ITERM_SESSION_ID", "w0t0p0:some-id")
	caps := DetectCapabilities()

	if caps.Colors != ColorTrue {
		t.Errorf("ITERM_SESSION_ID set: Colors = %v, want ColorTrue", caps.Colors)
	}
	if !caps.TrueColor {
		t.Error("ITERM_SESSION_ID should enable TrueColor")
	}
}

func TestDetectCapabilities_Kitty(t *testing.T) {
	env := newTestEnvHelper()
	defer env.Restore()
	clearTermEnvVars(env)

	env.Set("KITTY_WINDOW_ID", "1")
	caps := DetectCapabilities()

	if caps.Colors != ColorTrue {
		t.Errorf("KITTY_WINDOW_ID set: Colors = %v, want ColorTrue", caps.Colors)
	}
	if !caps.TrueColor {
		t.Error("KITTY_WINDOW_ID should enable TrueColor")
	}
}

func TestDetectCapabilities_Konsole(t *testing.T) {
	env := newTestEnvHelper()
	defer env.Restore()
	clearTermEnvVars(env)

	env.Set("KONSOLE_VERSION", "221201")
	caps := DetectCapabilities()

	if caps.Colors != ColorTrue {
		t.Errorf("KONSOLE_VERSION set: Colors = %v, want ColorTrue", caps.Colors)
	}
	if !caps.TrueColor {
		t.Error("KONSOLE_VERSION should enable TrueColor")
	}
}

func TestDetectCapabilities_VTE(t *testing.T) {
	env := newTestEnvHelper()
	defer env.Restore()
	clearTermEnvVars(env)

	env.Set("VTE_VERSION", "6800")
	caps := DetectCapabilities()

	if caps.Colors != ColorTrue {
		t.Errorf("VTE_VERSION set: Colors = %v, want ColorTrue", caps.Colors)
	}
	if !caps.TrueColor {
		t.Error("VTE_VERSION should enable TrueColor")
	}
}

func TestDetectCapabilities_DefaultMinimal(t *testing.T) {
	env := newTestEnvHelper()
	defer env.Restore()
	clearTermEnvVars(env)

	caps := DetectCapabilities()

	// With no env vars, should get safe defaults
	if caps.Colors != Color16 {
		t.Errorf("No env vars: Colors = %v, want Color16 (default)", caps.Colors)
	}
	if !caps.Unicode {
		t.Error("Unicode should be enabled by default")
	}
	if !caps.AltScreen {
		t.Error("AltScreen should be enabled by default")
	}
}

func TestCapabilities_SupportsColor(t *testing.T) {
	type tc struct {
		caps     Capabilities
		color    Color
		expected bool
	}

	tests := map[string]tc{
		"default color always supported": {
			caps:     Capabilities{Colors: ColorNone},
			color:    DefaultColor(),
			expected: true,
		},
		"ANSI color with Color16": {
			caps:     Capabilities{Colors: Color16},
			color:    ANSIColor(1),
			expected: true,
		},
		"ANSI color with ColorNone": {
			caps:     Capabilities{Colors: ColorNone},
			color:    ANSIColor(1),
			expected: false,
		},
		"RGB color with TrueColor": {
			caps:     Capabilities{Colors: ColorTrue, TrueColor: true},
			color:    RGBColor(255, 0, 0),
			expected: true,
		},
		"RGB color without TrueColor": {
			caps:     Capabilities{Colors: Color256, TrueColor: false},
			color:    RGBColor(255, 0, 0),
			expected: false,
		},
		"ANSI color with Color256": {
			caps:     Capabilities{Colors: Color256},
			color:    ANSIColor(200),
			expected: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.caps.SupportsColor(tt.color)
			if got != tt.expected {
				t.Errorf("SupportsColor() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCapabilities_EffectiveColor(t *testing.T) {
	type tc struct {
		caps         Capabilities
		color        Color
		expectedType ColorType
	}

	tests := map[string]tc{
		"RGB supported returns original": {
			caps:         Capabilities{Colors: ColorTrue, TrueColor: true},
			color:        RGBColor(255, 0, 0),
			expectedType: ColorRGB,
		},
		"RGB unsupported returns ANSI": {
			caps:         Capabilities{Colors: Color256, TrueColor: false},
			color:        RGBColor(255, 0, 0),
			expectedType: ColorANSI,
		},
		"RGB with ColorNone returns default": {
			caps:         Capabilities{Colors: ColorNone},
			color:        RGBColor(255, 0, 0),
			expectedType: ColorDefault,
		},
		"ANSI supported returns original": {
			caps:         Capabilities{Colors: Color16},
			color:        ANSIColor(1),
			expectedType: ColorANSI,
		},
		"ANSI with ColorNone returns default": {
			caps:         Capabilities{Colors: ColorNone},
			color:        ANSIColor(1),
			expectedType: ColorDefault,
		},
		"default always returns default": {
			caps:         Capabilities{Colors: ColorNone},
			color:        DefaultColor(),
			expectedType: ColorDefault,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.caps.EffectiveColor(tt.color)
			if got.Type() != tt.expectedType {
				t.Errorf("EffectiveColor().Type() = %v, want %v", got.Type(), tt.expectedType)
			}
		})
	}
}

func TestCapabilities_EffectiveColor_Approximation(t *testing.T) {
	caps := Capabilities{Colors: Color256, TrueColor: false}

	// Pure red RGB should approximate to ANSI red in the color cube
	color := RGBColor(255, 0, 0)
	effective := caps.EffectiveColor(color)

	if effective.Type() != ColorANSI {
		t.Fatalf("EffectiveColor() type = %v, want ColorANSI", effective.Type())
	}

	// Should be in the color cube (index 16+)
	if effective.ANSI() < 16 {
		t.Errorf("Expected color cube index, got %d", effective.ANSI())
	}
}

func TestCapabilities_String(t *testing.T) {
	type tc struct {
		caps     Capabilities
		contains []string
	}

	tests := map[string]tc{
		"true color with all features": {
			caps:     Capabilities{Colors: ColorTrue, Unicode: true, AltScreen: true},
			contains: []string{"true-color", "unicode", "altscreen"},
		},
		"256 color": {
			caps:     Capabilities{Colors: Color256, Unicode: true, AltScreen: true},
			contains: []string{"256-color"},
		},
		"16 color": {
			caps:     Capabilities{Colors: Color16, Unicode: true, AltScreen: true},
			contains: []string{"16-color"},
		},
		"no color": {
			caps:     Capabilities{Colors: ColorNone, Unicode: false, AltScreen: false},
			contains: []string{"no-color", "ascii"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			s := tt.caps.String()
			for _, substr := range tt.contains {
				if !contains(s, substr) {
					t.Errorf("String() = %q, should contain %q", s, substr)
				}
			}
		})
	}
}

// contains is a simple helper to check if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
