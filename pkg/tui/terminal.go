package tui

// ColorCapability describes the level of color support in a terminal.
type ColorCapability int

const (
	// ColorNone indicates a monochrome terminal with no color support.
	ColorNone ColorCapability = iota
	// Color16 indicates basic 16-color support (ANSI standard colors).
	Color16
	// Color256 indicates ANSI 256 palette support.
	Color256
	// ColorTrue indicates 24-bit true color (RGB) support.
	ColorTrue
)

// Capabilities describes what features the terminal supports.
type Capabilities struct {
	// Colors indicates the level of color support.
	Colors ColorCapability
	// Unicode indicates whether the terminal can render Unicode characters.
	Unicode bool
	// TrueColor indicates whether 24-bit RGB colors are supported.
	TrueColor bool
	// AltScreen indicates whether the terminal supports alternate screen buffer.
	AltScreen bool
}

// Terminal abstracts terminal operations for rendering and input.
// Implementations handle ANSI, Windows Console, or mock terminals for testing.
type Terminal interface {
	// Size returns the terminal dimensions (width, height) in cells.
	Size() (width, height int)

	// Flush writes the given cell changes to the terminal.
	// Changes are expected to be in row-major order for optimal performance.
	Flush(changes []CellChange)

	// Clear clears the entire terminal screen.
	Clear()

	// SetCursor moves the cursor to the specified position (0-indexed).
	SetCursor(x, y int)

	// HideCursor makes the cursor invisible.
	HideCursor()

	// ShowCursor makes the cursor visible.
	ShowCursor()

	// EnterRawMode puts the terminal into raw mode for character-by-character input.
	// Returns an error if raw mode cannot be enabled.
	EnterRawMode() error

	// ExitRawMode restores the terminal to its previous mode.
	// Returns an error if the previous mode cannot be restored.
	ExitRawMode() error

	// EnterAltScreen switches to the alternate screen buffer.
	// This preserves the original terminal content.
	EnterAltScreen()

	// ExitAltScreen switches back to the main screen buffer.
	// This restores the original terminal content.
	ExitAltScreen()

	// Caps returns the terminal's capabilities.
	Caps() Capabilities
}
