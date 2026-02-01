// Package tui provides terminal rendering primitives for building terminal user interfaces.
package tui

import (
	"errors"
	"math"
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

// ansi16RGB maps ANSI colors 0-15 to approximate RGB values.
// These are typical terminal color values; actual values vary by terminal.
var ansi16RGB = [16][3]uint8{
	{0, 0, 0},       // 0: Black
	{205, 49, 49},   // 1: Red
	{13, 188, 121},  // 2: Green
	{229, 229, 16},  // 3: Yellow
	{36, 114, 200},  // 4: Blue
	{188, 63, 188},  // 5: Magenta
	{17, 168, 205},  // 6: Cyan
	{229, 229, 229}, // 7: White
	{102, 102, 102}, // 8: Bright Black (Gray)
	{241, 76, 76},   // 9: Bright Red
	{35, 209, 139},  // 10: Bright Green
	{245, 245, 67},  // 11: Bright Yellow
	{59, 142, 234},  // 12: Bright Blue
	{214, 112, 214}, // 13: Bright Magenta
	{41, 184, 219},  // 14: Bright Cyan
	{255, 255, 255}, // 15: Bright White
}

// ToRGBValues returns the red, green, and blue components of any color.
// For ANSI colors, it approximates the RGB values.
// For default colors, it returns (0, 0, 0).
func (c Color) ToRGBValues() (r, g, b uint8) {
	switch c.typ {
	case ColorDefault:
		return 0, 0, 0
	case ColorRGB:
		return c.r, c.g, c.b
	case ColorANSI:
		idx := c.r
		if idx < 16 {
			// Standard ANSI colors 0-15
			rgb := ansi16RGB[idx]
			return rgb[0], rgb[1], rgb[2]
		} else if idx < 232 {
			// 6x6x6 color cube (indices 16-231)
			// index = 16 + 36*r + 6*g + b where r,g,b are 0-5
			idx -= 16
			ri := idx / 36
			gi := (idx % 36) / 6
			bi := idx % 6
			// Convert 0-5 to RGB: 0→0, 1→95, 2→135, 3→175, 4→215, 5→255
			cubeToRGB := func(v uint8) uint8 {
				if v == 0 {
					return 0
				}
				return 55 + v*40
			}
			return cubeToRGB(ri), cubeToRGB(gi), cubeToRGB(bi)
		} else {
			// Grayscale (indices 232-255)
			// 24 shades from dark gray to light gray
			gray := 8 + (idx-232)*10
			return gray, gray, gray
		}
	}
	return 0, 0, 0
}

// Luminance returns the relative luminance of the color (0.0-1.0).
// Uses the W3C formula for calculating relative luminance.
func (c Color) Luminance() float64 {
	if c.typ == ColorDefault {
		// Default color luminance is unknown; assume dark background
		return 0.0
	}
	r, g, b := c.ToRGBValues()

	// Convert to linear RGB (sRGB gamma correction)
	linearize := func(v uint8) float64 {
		f := float64(v) / 255.0
		if f <= 0.03928 {
			return f / 12.92
		}
		return math.Pow((f+0.055)/1.055, 2.4)
	}

	rLin := linearize(r)
	gLin := linearize(g)
	bLin := linearize(b)

	// W3C relative luminance formula
	return 0.2126*rLin + 0.7152*gLin + 0.0722*bLin
}

// IsLight returns true if the color is perceptually light.
// Uses a luminance threshold of 0.5 (middle gray).
func (c Color) IsLight() bool {
	if c.typ == ColorDefault {
		return false // Assume default is dark
	}
	return c.Luminance() > 0.2
}
