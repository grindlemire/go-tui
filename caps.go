package tui

import (
	"os"
	"strings"
)

// DetectCapabilities determines terminal capabilities from environment variables
// and returns a Capabilities struct with detected settings.
// Returns conservative defaults when detection fails.
func DetectCapabilities() Capabilities {
	caps := Capabilities{
		Colors:    Color16,  // Safe default for most terminals
		Unicode:   true,     // Assume modern terminal
		TrueColor: false,
		AltScreen: true,
	}

	// First, check for explicit true color indicators that override everything else.
	// These environment variables definitively indicate true color support.

	// Check COLORTERM for explicit true color support
	colorterm := strings.ToLower(os.Getenv("COLORTERM"))
	if colorterm == "truecolor" || colorterm == "24bit" {
		caps.Colors = ColorTrue
		caps.TrueColor = true
	}

	// Check terminal emulator-specific environment variables
	// These terminals are known to support true color

	// Windows Terminal
	if os.Getenv("WT_SESSION") != "" {
		caps.Colors = ColorTrue
		caps.TrueColor = true
	}

	// iTerm2
	if os.Getenv("ITERM_SESSION_ID") != "" {
		caps.Colors = ColorTrue
		caps.TrueColor = true
	}

	// Kitty
	if os.Getenv("KITTY_WINDOW_ID") != "" {
		caps.Colors = ColorTrue
		caps.TrueColor = true
	}

	// Konsole
	if os.Getenv("KONSOLE_VERSION") != "" {
		caps.Colors = ColorTrue
		caps.TrueColor = true
	}

	// VTE-based terminals (GNOME Terminal, Tilix, etc.)
	if os.Getenv("VTE_VERSION") != "" {
		caps.Colors = ColorTrue
		caps.TrueColor = true
	}

	// If we already detected true color via explicit indicators, we're done
	if caps.TrueColor {
		return caps
	}

	// Now check TERM environment variable for terminals without explicit indicators
	term := strings.ToLower(os.Getenv("TERM"))
	switch {
	case term == "dumb":
		caps.Colors = ColorNone
		caps.Unicode = false
		caps.AltScreen = false
		return caps // Early return for truly dumb terminal
	case strings.Contains(term, "256color"):
		caps.Colors = Color256
	case strings.Contains(term, "truecolor"):
		caps.Colors = ColorTrue
		caps.TrueColor = true
	}

	return caps
}

// SupportsColor returns true if the terminal supports the given color type.
func (c Capabilities) SupportsColor(color Color) bool {
	switch color.Type() {
	case ColorDefault:
		return true
	case ColorANSI:
		// ANSI colors need at least Color16 support
		return c.Colors >= Color16
	case ColorRGB:
		return c.TrueColor
	}
	return false
}

// EffectiveColor returns the color to use given the terminal's capabilities.
// If the terminal supports the color type, returns the original color.
// If not, returns an appropriate fallback (RGB colors are approximated to ANSI).
func (c Capabilities) EffectiveColor(color Color) Color {
	if c.SupportsColor(color) {
		return color
	}

	switch color.Type() {
	case ColorRGB:
		// Fall back to ANSI approximation
		if c.Colors >= Color256 {
			return color.ToANSI()
		}
		// For 16-color terminals, we could add further approximation
		// For now, return the ANSI approximation which uses the 256 palette
		if c.Colors >= Color16 {
			return color.ToANSI()
		}
		// No color support - return default
		return DefaultColor()
	case ColorANSI:
		// If ANSI not supported, return default
		if c.Colors < Color16 {
			return DefaultColor()
		}
		return color
	default:
		return color
	}
}

// String returns a human-readable description of the capabilities.
func (c Capabilities) String() string {
	var parts []string

	switch c.Colors {
	case ColorNone:
		parts = append(parts, "no-color")
	case Color16:
		parts = append(parts, "16-color")
	case Color256:
		parts = append(parts, "256-color")
	case ColorTrue:
		parts = append(parts, "true-color")
	}

	if c.Unicode {
		parts = append(parts, "unicode")
	} else {
		parts = append(parts, "ascii")
	}

	if c.AltScreen {
		parts = append(parts, "altscreen")
	}

	return strings.Join(parts, ", ")
}
