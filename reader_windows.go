//go:build windows

package tui

import (
	"os"
	"time"
)

// stdinReader implements EventReader for Windows terminals.
// Resize delivery is not yet wired on Windows; key/mouse decoding still uses parseInput.
type stdinReader struct {
	in         *os.File
	buf        []byte
	pending    []Event
	interruptC chan struct{}
}

var _ InterruptibleReader = (*stdinReader)(nil)

// NewEventReader creates an EventReader for the given terminal input.
func NewEventReader(in *os.File) (EventReader, error) {
	return &stdinReader{
		in:  in,
		buf: make([]byte, 256),
	}, nil
}

func (r *stdinReader) PollEvent(timeout time.Duration) (Event, bool) {
	if len(r.pending) > 0 {
		ev := r.pending[0]
		r.pending = r.pending[1:]
		return ev, true
	}

	if timeout == 0 {
		return nil, false
	}

	type readResult struct {
		n   int
		err error
	}
	done := make(chan readResult, 1)
	go func() {
		n, err := r.in.Read(r.buf)
		done <- readResult{n: n, err: err}
	}()

	if timeout < 0 {
		select {
		case <-r.interruptC:
			return nil, false
		case res := <-done:
			if res.err != nil || res.n == 0 {
				return nil, false
			}
			r.pending = parseInput(r.buf[:res.n])
		}
	} else {
		timer := time.NewTimer(timeout)
		defer timer.Stop()
		select {
		case <-r.interruptC:
			return nil, false
		case <-timer.C:
			return nil, false
		case res := <-done:
			if res.err != nil || res.n == 0 {
				return nil, false
			}
			r.pending = parseInput(r.buf[:res.n])
		}
	}

	if len(r.pending) == 0 {
		return nil, false
	}
	ev := r.pending[0]
	r.pending = r.pending[1:]
	return ev, true
}

func (r *stdinReader) Close() error {
	return nil
}

func (r *stdinReader) EnableInterrupt() error {
	if r.interruptC == nil {
		r.interruptC = make(chan struct{}, 1)
	}
	return nil
}

func (r *stdinReader) Interrupt() error {
	if r.interruptC == nil {
		return nil
	}
	select {
	case r.interruptC <- struct{}{}:
	default:
	}
	return nil
}
