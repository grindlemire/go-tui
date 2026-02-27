package tui

import (
	"io"
	"sync/atomic"
)

// inlineStreamWriter implements io.WriteCloser for streaming text to the
// inline history region. Writes are queued onto the app's main event loop.
// Goroutine-safe.
type inlineStreamWriter struct {
	app    *App
	closed atomic.Bool
}

func newInlineStreamWriter(app *App) *inlineStreamWriter {
	return &inlineStreamWriter{app: app}
}

func (w *inlineStreamWriter) Write(p []byte) (int, error) {
	if w.closed.Load() {
		return 0, io.ErrClosedPipe
	}
	// Copy data so the caller can reuse the slice.
	data := make([]byte, len(p))
	copy(data, p)

	w.app.QueueUpdate(func() {
		w.app.ensureInlineSession()
		w.app.inlineSession.ensureInitialized(&w.app.inlineLayout, w.app.inlineStartRow)
		width, _ := w.app.terminal.Size()
		w.app.inlineSession.appendBytes(&w.app.inlineLayout, w.app.inlineStartRow, width, data)
		w.app.MarkDirty()
	})
	return len(p), nil
}

func (w *inlineStreamWriter) Close() error {
	if w.closed.Swap(true) {
		return nil // already closed
	}
	w.app.QueueUpdate(func() {
		w.app.ensureInlineSession()
		w.app.inlineSession.finalizePartial(&w.app.inlineLayout)
		if w.app.activeStreamWriter == w {
			w.app.activeStreamWriter = nil
		}
		w.app.MarkDirty()
	})
	return nil
}

// nopStreamWriter is returned when StreamAbove is called outside inline mode.
type nopStreamWriter struct{}

func (w *nopStreamWriter) Write(p []byte) (int, error) { return len(p), nil }
func (w *nopStreamWriter) Close() error                { return nil }
