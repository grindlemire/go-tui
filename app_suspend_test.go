//go:build !windows

package tui

// recordingTerminal wraps MockTerminal and records method calls in order.
type recordingTerminal struct {
	*MockTerminal
	calls []string
}

func newRecordingTerminal(width, height int) *recordingTerminal {
	return &recordingTerminal{
		MockTerminal: NewMockTerminal(width, height),
	}
}

func (r *recordingTerminal) DisableMouse() {
	r.calls = append(r.calls, "DisableMouse")
	r.MockTerminal.DisableMouse()
}

func (r *recordingTerminal) ShowCursor() {
	r.calls = append(r.calls, "ShowCursor")
	r.MockTerminal.ShowCursor()
}

func (r *recordingTerminal) HideCursor() {
	r.calls = append(r.calls, "HideCursor")
	r.MockTerminal.HideCursor()
}

func (r *recordingTerminal) ExitAltScreen() {
	r.calls = append(r.calls, "ExitAltScreen")
	r.MockTerminal.ExitAltScreen()
}

func (r *recordingTerminal) EnterAltScreen() {
	r.calls = append(r.calls, "EnterAltScreen")
	r.MockTerminal.EnterAltScreen()
}

func (r *recordingTerminal) ExitRawMode() error {
	r.calls = append(r.calls, "ExitRawMode")
	return r.MockTerminal.ExitRawMode()
}

func (r *recordingTerminal) EnterRawMode() error {
	r.calls = append(r.calls, "EnterRawMode")
	return r.MockTerminal.EnterRawMode()
}

func (r *recordingTerminal) EnableMouse() {
	r.calls = append(r.calls, "EnableMouse")
	r.MockTerminal.EnableMouse()
}

func (r *recordingTerminal) Clear() {
	r.calls = append(r.calls, "Clear")
	r.MockTerminal.Clear()
}
