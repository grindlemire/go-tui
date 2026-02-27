package tui

import (
	"fmt"
	"io"
	"testing"
)

func TestNopStreamWriter_WritesSucceed(t *testing.T) {
	w := &nopStreamWriter{}
	n, err := w.Write([]byte("hello"))
	if err != nil {
		t.Fatalf("Write error: %v", err)
	}
	if n != 5 {
		t.Fatalf("Write n = %d, want 5", n)
	}
}

func TestNopStreamWriter_CloseSucceeds(t *testing.T) {
	w := &nopStreamWriter{}
	if err := w.Close(); err != nil {
		t.Fatalf("Close error: %v", err)
	}
}

func TestInlineStreamWriter_WriteAfterClose(t *testing.T) {
	app, _ := newInlineTestApp(80, 24, 3)
	w := newInlineStreamWriter(app)

	_ = w.Close()
	runQueuedUpdates(app)

	_, err := w.Write([]byte("after close"))
	if err != io.ErrClosedPipe {
		t.Fatalf("Write after Close: err = %v, want io.ErrClosedPipe", err)
	}
}

func TestInlineStreamWriter_DoubleCloseNoop(t *testing.T) {
	app, _ := newInlineTestApp(80, 24, 3)
	w := newInlineStreamWriter(app)

	if err := w.Close(); err != nil {
		t.Fatalf("first Close: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}
}

func TestInlineStreamWriter_StreamsToHistory(t *testing.T) {
	app, emu := newInlineTestApp(80, 24, 3)
	w := newInlineStreamWriter(app)

	fmt.Fprint(w, "streaming")
	runQueuedUpdates(app)

	historyBottom := app.inlineStartRow - 1
	got := emu.ScreenRow(historyBottom)
	if got != "streaming" {
		t.Fatalf("history row = %q, want %q\n%s", got, "streaming", emu.DumpState())
	}

	_ = w.Close()
}

func TestInlineStreamWriter_NewlineFinalizesRow(t *testing.T) {
	app, emu := newInlineTestApp(80, 24, 3)
	w := newInlineStreamWriter(app)

	fmt.Fprintln(w, "line1")
	fmt.Fprint(w, "line2")
	runQueuedUpdates(app)

	if got := emu.ScreenRow(app.inlineStartRow - 2); got != "line1" {
		t.Fatalf("row -2 = %q, want %q\n%s", got, "line1", emu.DumpState())
	}
	if got := emu.ScreenRow(app.inlineStartRow - 1); got != "line2" {
		t.Fatalf("row -1 = %q, want %q\n%s", got, "line2", emu.DumpState())
	}

	_ = w.Close()
}
