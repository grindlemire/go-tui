//go:build windows

package tui

import (
	"os"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modkernel32           = windows.NewLazySystemDLL("kernel32.dll")
	procReadConsoleInputW = modkernel32.NewProc("ReadConsoleInputW")
)

const (
	infiniteWait   = 0xFFFFFFFF
	maxRecordsRead = 64
)

type coord struct {
	X, Y int16
}

// inputRecord mirrors INPUT_RECORD. EventType (WORD) is followed by 2 bytes of
// padding before the 16-byte union of event-specific records. Total: 20 bytes.
type inputRecord struct {
	eventType uint16
	_         uint16
	data      [16]byte
}

type keyEventRecord struct {
	keyDown         int32
	repeatCount     uint16
	virtualKeyCode  uint16
	virtualScanCode uint16
	unicodeChar     uint16
	controlKeyState uint32
}

type mouseEventRecord struct {
	mousePosition   coord
	buttonState     uint32
	controlKeyState uint32
	eventFlags      uint32
}

type windowBufferSizeRecord struct {
	size coord
}

// readConsoleInputW wraps the ReadConsoleInputW syscall, which is not exposed
// by golang.org/x/sys/windows. Returns the number of records read.
func readConsoleInputW(h windows.Handle, recs []inputRecord) (int, error) {
	if len(recs) == 0 {
		return 0, nil
	}
	var read uint32
	r1, _, e1 := syscall.SyscallN(
		procReadConsoleInputW.Addr(),
		uintptr(h),
		uintptr(unsafe.Pointer(&recs[0])),
		uintptr(len(recs)),
		uintptr(unsafe.Pointer(&read)),
	)
	if r1 == 0 {
		if e1 != 0 {
			return 0, e1
		}
		return 0, syscall.EINVAL
	}
	return int(read), nil
}

// consoleViewportSize returns the current console viewport dimensions, read
// from the screen buffer info's window rect. Returns ok=false if the syscall
// fails or the stdout handle is not a console.
func consoleViewportSize() (width, height int, ok bool) {
	h, err := windows.GetStdHandle(windows.STD_OUTPUT_HANDLE)
	if err != nil || h == 0 {
		return 0, 0, false
	}
	var info windows.ConsoleScreenBufferInfo
	if err := windows.GetConsoleScreenBufferInfo(h, &info); err != nil {
		return 0, 0, false
	}
	return int(info.Window.Right - info.Window.Left + 1),
		int(info.Window.Bottom - info.Window.Top + 1),
		true
}

// stdinReader implements EventReader for Windows terminals by reading raw
// INPUT_RECORDs from the console input handle. Resize, key, and mouse events
// are decoded directly without going through parseInput.
type stdinReader struct {
	handle    windows.Handle
	interrupt windows.Handle
	pending   []Event
	paused    atomic.Bool

	// Latched button state from the last MOUSE_EVENT_RECORD. Diffing against
	// the previous record yields press/release; non-zero state during a
	// MOUSE_MOVED event indicates a drag.
	lastMouseButtons uint32
}

var (
	_ InterruptibleReader = (*stdinReader)(nil)
	_ PausableReader      = (*stdinReader)(nil)
)

// NewEventReader creates an EventReader for the given terminal input.
func NewEventReader(in *os.File) (EventReader, error) {
	return &stdinReader{
		handle: windows.Handle(in.Fd()),
	}, nil
}

func (r *stdinReader) Pause() {
	r.paused.Store(true)
	if r.interrupt != 0 {
		_ = r.Interrupt()
	}
}

func (r *stdinReader) Resume() {
	r.paused.Store(false)
}

// PollEvent reads the next event with a timeout. A negative timeout blocks
// until input arrives or Interrupt() is called.
func (r *stdinReader) PollEvent(timeout time.Duration) (Event, bool) {
	if r.paused.Load() {
		return nil, false
	}
	if len(r.pending) > 0 {
		ev := r.pending[0]
		r.pending = r.pending[1:]
		return ev, true
	}

	handles := []windows.Handle{r.handle}
	if r.interrupt != 0 {
		handles = append(handles, r.interrupt)
	}

	waited, err := windows.WaitForMultipleObjects(handles, false, timeoutToMs(timeout))
	if err != nil {
		return nil, false
	}
	switch waited {
	case uint32(windows.WAIT_TIMEOUT):
		return nil, false
	case windows.WAIT_OBJECT_0:
		// stdin signaled; fall through to read records.
	case windows.WAIT_OBJECT_0 + 1:
		// Interrupt fired. Reset for next interrupt.
		_ = windows.ResetEvent(r.interrupt)
		return nil, false
	default:
		return nil, false
	}

	// Drain however many records are queued, capped to avoid unbounded reads.
	var available uint32
	if err := windows.GetNumberOfConsoleInputEvents(r.handle, &available); err != nil || available == 0 {
		return nil, false
	}
	if available > maxRecordsRead {
		available = maxRecordsRead
	}
	recs := make([]inputRecord, available)
	n, err := readConsoleInputW(r.handle, recs)
	if err != nil || n == 0 {
		return nil, false
	}
	for i := range n {
		r.decodeRecord(&recs[i])
	}

	if len(r.pending) == 0 {
		return nil, false
	}
	ev := r.pending[0]
	r.pending = r.pending[1:]
	return ev, true
}

func (r *stdinReader) Close() error {
	if r.interrupt != 0 {
		_ = windows.CloseHandle(r.interrupt)
		r.interrupt = 0
	}
	return nil
}

// EnableInterrupt creates a manual-reset Win32 event that, when signaled,
// wakes a blocked PollEvent via WaitForMultipleObjects.
func (r *stdinReader) EnableInterrupt() error {
	if r.interrupt != 0 {
		return nil
	}
	h, err := windows.CreateEvent(nil, 1, 0, nil)
	if err != nil {
		return err
	}
	r.interrupt = h
	return nil
}

func (r *stdinReader) Interrupt() error {
	if r.interrupt == 0 {
		return nil
	}
	return windows.SetEvent(r.interrupt)
}

func timeoutToMs(timeout time.Duration) uint32 {
	switch {
	case timeout < 0:
		return infiniteWait
	case timeout == 0:
		return 0
	}
	ms := timeout.Milliseconds()
	if ms <= 0 {
		// Sub-millisecond positive timeout should not become a 0 (non-blocking) wait.
		return 1
	}
	if ms > int64(infiniteWait-1) {
		return infiniteWait - 1
	}
	return uint32(ms)
}

func (r *stdinReader) decodeRecord(rec *inputRecord) {
	switch rec.eventType {
	case windows.KEY_EVENT:
		kev := (*keyEventRecord)(unsafe.Pointer(&rec.data))
		if kev.keyDown == 0 {
			return
		}
		repeats := max(1, int(kev.repeatCount))
		ev, ok := translateKeyEvent(kev)
		if !ok {
			return
		}
		for range repeats {
			r.pending = append(r.pending, ev)
		}
	case windows.MOUSE_EVENT:
		mev := (*mouseEventRecord)(unsafe.Pointer(&rec.data))
		if ev, ok := r.translateMouseEvent(mev); ok {
			r.pending = append(r.pending, ev)
		}
	case windows.WINDOW_BUFFER_SIZE_EVENT:
		// dwSize on the record reports the *buffer* dimensions, which on
		// legacy cmd.exe can be larger than the visible viewport. Read the
		// viewport rect from the screen buffer info to get the true visible
		// area, matching terminal_windows.go's getTerminalSize.
		if w, h, ok := consoleViewportSize(); ok {
			r.pending = append(r.pending, ResizeEvent{Width: w, Height: h})
		} else {
			// Fall back to dwSize if the syscall fails. Better an approximate
			// resize event than none at all.
			wev := (*windowBufferSizeRecord)(unsafe.Pointer(&rec.data))
			r.pending = append(r.pending, ResizeEvent{
				Width:  int(wev.size.X),
				Height: int(wev.size.Y),
			})
		}
	}
}

// translateKeyEvent converts a KEY_EVENT_RECORD into a tui.KeyEvent, matching
// the legacy-protocol normalization used by parseInput on Unix where possible.
func translateKeyEvent(kev *keyEventRecord) (KeyEvent, bool) {
	mod := windowsKeyMod(kev.controlKeyState)

	if k, ok := vkToKey(kev.virtualKeyCode); ok {
		return KeyEvent{Key: k, Mod: mod}, true
	}

	r := rune(kev.unicodeChar)
	if r == 0 {
		// Pure modifier-only key event (Shift, Ctrl alone, etc.). Drop it.
		return KeyEvent{}, false
	}

	// Control bytes: normalize 0x01..0x1A to KeyRune 'a'..'z' with ModCtrl so
	// app-level handlers can use Rune('a').Ctrl() the same way they do on Unix.
	if r > 0 && r < 0x20 {
		switch r {
		case 0x08:
			return KeyEvent{Key: KeyBackspace, Mod: mod &^ ModCtrl}, true
		case 0x09:
			return KeyEvent{Key: KeyTab, Mod: mod &^ ModCtrl}, true
		case 0x0d:
			return KeyEvent{Key: KeyEnter, Mod: mod &^ ModCtrl}, true
		case 0x1b:
			return KeyEvent{Key: KeyEscape, Mod: mod &^ ModCtrl}, true
		}
		return KeyEvent{Key: KeyRune, Rune: rune('a' + r - 1), Mod: mod | ModCtrl}, true
	}
	if r == 0x7f {
		return KeyEvent{Key: KeyBackspace, Mod: mod}, true
	}
	// Printable rune. The character itself already encodes the Shift state
	// ('A' vs 'a'), so drop ModShift to match the Unix behavior.
	return KeyEvent{Key: KeyRune, Rune: r, Mod: mod &^ ModShift}, true
}

func vkToKey(vk uint16) (Key, bool) {
	switch vk {
	case windows.VK_ESCAPE:
		return KeyEscape, true
	case windows.VK_RETURN:
		return KeyEnter, true
	case windows.VK_TAB:
		return KeyTab, true
	case windows.VK_BACK:
		return KeyBackspace, true
	case windows.VK_DELETE:
		return KeyDelete, true
	case windows.VK_INSERT:
		return KeyInsert, true
	case windows.VK_UP:
		return KeyUp, true
	case windows.VK_DOWN:
		return KeyDown, true
	case windows.VK_LEFT:
		return KeyLeft, true
	case windows.VK_RIGHT:
		return KeyRight, true
	case windows.VK_HOME:
		return KeyHome, true
	case windows.VK_END:
		return KeyEnd, true
	case windows.VK_PRIOR:
		return KeyPageUp, true
	case windows.VK_NEXT:
		return KeyPageDown, true
	case windows.VK_F1:
		return KeyF1, true
	case windows.VK_F1 + 1:
		return KeyF2, true
	case windows.VK_F1 + 2:
		return KeyF3, true
	case windows.VK_F1 + 3:
		return KeyF4, true
	case windows.VK_F1 + 4:
		return KeyF5, true
	case windows.VK_F1 + 5:
		return KeyF6, true
	case windows.VK_F1 + 6:
		return KeyF7, true
	case windows.VK_F1 + 7:
		return KeyF8, true
	case windows.VK_F1 + 8:
		return KeyF9, true
	case windows.VK_F1 + 9:
		return KeyF10, true
	case windows.VK_F1 + 10:
		return KeyF11, true
	case windows.VK_F12:
		return KeyF12, true
	}
	return KeyNone, false
}

func windowsKeyMod(state uint32) Modifier {
	var m Modifier
	if state&(windows.LEFT_CTRL_PRESSED|windows.RIGHT_CTRL_PRESSED) != 0 {
		m |= ModCtrl
	}
	if state&(windows.LEFT_ALT_PRESSED|windows.RIGHT_ALT_PRESSED) != 0 {
		m |= ModAlt
	}
	if state&windows.SHIFT_PRESSED != 0 {
		m |= ModShift
	}
	return m
}

// translateMouseEvent converts a MOUSE_EVENT_RECORD into a MouseEvent. Press
// and release are derived by diffing against the previously latched button
// state. Pure hover motion (no buttons held) is dropped.
func (r *stdinReader) translateMouseEvent(mev *mouseEventRecord) (MouseEvent, bool) {
	ev := MouseEvent{
		X:   int(mev.mousePosition.X),
		Y:   int(mev.mousePosition.Y),
		Mod: windowsKeyMod(mev.controlKeyState),
	}

	switch {
	case mev.eventFlags&windows.MOUSE_WHEELED != 0:
		// High word of buttonState is the signed wheel delta; positive = up.
		if int16(mev.buttonState>>16) > 0 {
			ev.Button = MouseWheelUp
		} else {
			ev.Button = MouseWheelDown
		}
		ev.Action = MousePress
		return ev, true

	case mev.eventFlags&windows.MOUSE_MOVED != 0:
		if r.lastMouseButtons == 0 {
			return MouseEvent{}, false
		}
		ev.Button = buttonFromState(r.lastMouseButtons)
		ev.Action = MouseDrag
		return ev, true

	default:
		prev := r.lastMouseButtons
		curr := mev.buttonState
		r.lastMouseButtons = curr
		if pressed := curr &^ prev; pressed != 0 {
			ev.Button = buttonFromState(pressed)
			ev.Action = MousePress
			return ev, true
		}
		if released := prev &^ curr; released != 0 {
			ev.Button = buttonFromState(released)
			ev.Action = MouseRelease
			return ev, true
		}
		return MouseEvent{}, false
	}
}

func buttonFromState(state uint32) MouseButton {
	switch {
	case state&windows.FROM_LEFT_1ST_BUTTON_PRESSED != 0:
		return MouseLeft
	case state&windows.RIGHTMOST_BUTTON_PRESSED != 0:
		return MouseRight
	case state&windows.FROM_LEFT_2ND_BUTTON_PRESSED != 0:
		return MouseMiddle
	}
	return MouseNone
}
