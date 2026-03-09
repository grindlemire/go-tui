//go:build !windows

package tui

import (
	"os"
	"os/signal"
	"syscall"
)

// registerResizeSignal sets up a SIGWINCH handler to dispatch resize events
// when the terminal window is resized. Returns a cleanup function.
func (a *App) registerResizeSignal() func() {
	resizeCh := make(chan os.Signal, 10)
	signal.Notify(resizeCh, syscall.SIGWINCH)

	go func() {
		for {
			select {
			case <-resizeCh:
				w, h := a.terminal.Size()
				ev := ResizeEvent{Width: w, Height: h}
				// Wake blocking reader so the resize event is processed promptly
				if interruptible, ok := a.reader.(InterruptibleReader); ok {
					interruptible.Interrupt()
				}
				select {
				case a.eventQueue <- func() { a.Dispatch(ev) }:
				case <-a.stopCh:
					return
				}
			case <-a.stopCh:
				return
			}
		}
	}()

	return func() {
		signal.Stop(resizeCh)
	}
}
