//go:build windows

package tui

import "golang.org/x/sys/windows"

// rawModeState stores the original console mode for restoration.
type rawModeState struct {
	fd   windows.Handle
	mode uint32
}

// enableRawMode puts the Windows console into raw-ish mode and returns the
// previous mode for restoration.
func enableRawMode(fd int) (*rawModeState, error) {
	h := windows.Handle(fd)

	var mode uint32
	if err := windows.GetConsoleMode(h, &mode); err != nil {
		return nil, err
	}

	raw := mode
	raw &^= windows.ENABLE_ECHO_INPUT | windows.ENABLE_LINE_INPUT | windows.ENABLE_PROCESSED_INPUT
	raw |= windows.ENABLE_EXTENDED_FLAGS | windows.ENABLE_WINDOW_INPUT | windows.ENABLE_VIRTUAL_TERMINAL_INPUT

	if err := windows.SetConsoleMode(h, raw); err != nil {
		return nil, err
	}

	return &rawModeState{fd: h, mode: mode}, nil
}

// disableRawMode restores the console to its previous mode.
func disableRawMode(state *rawModeState) error {
	if state == nil {
		return nil
	}
	return windows.SetConsoleMode(state.fd, state.mode)
}

// getTerminalSize returns the terminal dimensions.
func getTerminalSize(fd int) (width, height int, err error) {
	h := windows.Handle(fd)
	var info windows.ConsoleScreenBufferInfo
	if err := windows.GetConsoleScreenBufferInfo(h, &info); err != nil {
		return 0, 0, err
	}

	width = int(info.Window.Right - info.Window.Left + 1)
	height = int(info.Window.Bottom - info.Window.Top + 1)
	return width, height, nil
}
