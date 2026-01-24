// Package tui provides terminal rendering primitives for building terminal user interfaces.
package tui

import (
	"errors"
	"strings"
)

// ColorType distinguishes between color representations.
type ColorType uint8

const (
	// ColorDefault represents the terminal's default color (no color set).
	ColorDefault ColorType = iota
	// ColorANSI represents an ANSI 256 palette color (0-255).
	ColorANSI
	// ColorRGB represents a true color (24-bit RGB).
	ColorRGB
)

// Color represents a terminal color with support for default, ANSI 256, and true color.
// Zero value represents the terminal default color.
type Color struct {
	typ ColorType
	// For ANSI: r holds the palette index (0-255)
	// For RGB: r, g, b hold the color components
	r, g, b uint8
}

// DefaultColor returns a Color representing the terminal's default color.
func DefaultColor() Color {
	return Color{typ: ColorDefault}
}

// ANSIColor returns a Color from the ANSI 256 palette.
func ANSIColor(index uint8) Color {
	return Color{typ: ColorANSI, r: index}
}

// RGBColor returns a true color (24-bit RGB) Color.
func RGBColor(r, g, b uint8) Color {
	return Color{typ: ColorRGB, r: r, g: g, b: b}
}

// HexColor parses a hex color string and returns a Color.
// Supported formats: "#RRGGBB" and "#RGB".
func HexColor(hex string) (Color, error) {
	hex = strings.TrimPrefix(hex, "#")

	switch len(hex) {
	case 6:
		// #RRGGBB
		r, err := parseHexByte(hex[0:2])
		if err != nil {
			return Color{}, err
		}
		g, err := parseHexByte(hex[2:4])
		if err != nil {
			return Color{}, err
		}
		b, err := parseHexByte(hex[4:6])
		if err != nil {
			return Color{}, err
		}
		return RGBColor(r, g, b), nil
	case 3:
		// #RGB -> expand to #RRGGBB
		r, err := parseHexNibble(hex[0])
		if err != nil {
			return Color{}, err
		}
		g, err := parseHexNibble(hex[1])
		if err != nil {
			return Color{}, err
		}
		b, err := parseHexNibble(hex[2])
		if err != nil {
			return Color{}, err
		}
		// Expand nibble to byte: 0xF -> 0xFF
		return RGBColor(r<<4|r, g<<4|g, b<<4|b), nil
	default:
		return Color{}, errors.New("invalid hex color format: expected #RGB or #RRGGBB")
	}
}

// parseHexByte parses a two-character hex string into a byte.
func parseHexByte(s string) (uint8, error) {
	if len(s) != 2 {
		return 0, errors.New("invalid hex byte")
	}
	high, err := parseHexNibble(s[0])
	if err != nil {
		return 0, err
	}
	low, err := parseHexNibble(s[1])
	if err != nil {
		return 0, err
	}
	return high<<4 | low, nil
}

// parseHexNibble parses a single hex character into a nibble (0-15).
func parseHexNibble(c byte) (uint8, error) {
	switch {
	case c >= '0' && c <= '9':
		return c - '0', nil
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10, nil
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10, nil
	default:
		return 0, errors.New("invalid hex character")
	}
}

// Type returns the ColorType of this color.
func (c Color) Type() ColorType {
	return c.typ
}

// IsDefault returns true if this is the terminal's default color.
func (c Color) IsDefault() bool {
	return c.typ == ColorDefault
}

// ANSI returns the ANSI palette index.
// Panics if the color is not an ANSI color.
func (c Color) ANSI() uint8 {
	if c.typ != ColorANSI {
		panic("Color.ANSI() called on non-ANSI color")
	}
	return c.r
}

// RGB returns the red, green, and blue components.
// Panics if the color is not an RGB color.
func (c Color) RGB() (r, g, b uint8) {
	if c.typ != ColorRGB {
		panic("Color.RGB() called on non-RGB color")
	}
	return c.r, c.g, c.b
}

// Equal returns true if both colors are identical.
func (c Color) Equal(other Color) bool {
	if c.typ != other.typ {
		return false
	}
	switch c.typ {
	case ColorDefault:
		return true
	case ColorANSI:
		return c.r == other.r
	case ColorRGB:
		return c.r == other.r && c.g == other.g && c.b == other.b
	}
	return false
}

// ToANSI approximates an RGB color to the nearest ANSI 256 palette entry.
// Uses the 6x6x6 color cube (indices 16-231) plus grayscale (232-255).
// Returns the color unchanged if it's already ANSI or default.
func (c Color) ToANSI() Color {
	if c.typ != ColorRGB {
		return c
	}

	r, g, b := c.r, c.g, c.b

	// Check if grayscale (or close to it)
	if r == g && g == b {
		// Grayscale ramp: 232-255 (24 shades)
		// 0 maps to 232, 255 maps to 255
		if r < 8 {
			return ANSIColor(16) // Black in the color cube is closer
		}
		if r > 248 {
			return ANSIColor(231) // White in the color cube is closer
		}
		gray := uint8(232 + (int(r)-8)*24/240)
		return ANSIColor(gray)
	}

	// 6x6x6 color cube: 16-231
	// Each component maps to 0-5
	ri := int(r) * 5 / 255
	gi := int(g) * 5 / 255
	bi := int(b) * 5 / 255

	index := uint8(16 + 36*ri + 6*gi + bi)
	return ANSIColor(index)
}

// Standard ANSI colors (basic 8 colors).
var (
	Black   = ANSIColor(0)
	Red     = ANSIColor(1)
	Green   = ANSIColor(2)
	Yellow  = ANSIColor(3)
	Blue    = ANSIColor(4)
	Magenta = ANSIColor(5)
	Cyan    = ANSIColor(6)
	White   = ANSIColor(7)
)

// Bright ANSI colors (high-intensity variants).
var (
	BrightBlack   = ANSIColor(8)
	BrightRed     = ANSIColor(9)
	BrightGreen   = ANSIColor(10)
	BrightYellow  = ANSIColor(11)
	BrightBlue    = ANSIColor(12)
	BrightMagenta = ANSIColor(13)
	BrightCyan    = ANSIColor(14)
	BrightWhite   = ANSIColor(15)
)
