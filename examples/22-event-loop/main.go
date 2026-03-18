// Package main demonstrates three ways to drive go-tui's event loop.
//
// Each mode renders the same UI: a live message feed from a background
// producer. The difference is how the event loop is wired.
//
// Usage:
//
//	go run ../../cmd/tui generate feed.gsx
//	go run . run       # Standard Run() - go-tui owns the loop
//	go run . step      # Step() - you own the loop, call Step() each frame
//	go run . select    # Events() + select - multiplex with your own channels
package main

import (
	"fmt"
	"os"
	"time"

	tui "github.com/grindlemire/go-tui"
)

//go:generate go run ../../cmd/tui generate feed.gsx

func main() {
	mode := "run"
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}

	switch mode {
	case "run":
		runMode()
	case "step":
		stepMode()
	case "select":
		selectMode()
	default:
		fmt.Fprintf(os.Stderr, "Unknown mode %q. Use: run, step, or select\n", mode)
		os.Exit(1)
	}
}

// startProducer launches a goroutine that sends timestamped messages to a
// channel at ~5/sec. Checks paused() before sending.
func startProducer(paused func() bool) <-chan string {
	ch := make(chan string, 10)
	go func() {
		for i := 1; ; i++ {
			time.Sleep(200 * time.Millisecond)
			if paused() {
				continue
			}
			select {
			case ch <- fmt.Sprintf("[%s] Message #%d", time.Now().Format("15:04:05.000"), i):
			default:
			}
		}
	}()
	return ch
}

// runMode uses the standard Run() approach. The background producer pushes
// messages through QueueUpdate because it runs on a separate goroutine.
// This is the simplest approach and works for most apps.
func runMode() {
	comp := NewFeedApp("Run()")
	app, err := tui.NewApp(tui.WithRootComponent(comp))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer app.Close()

	// The producer must use QueueUpdate to safely mutate state from its
	// goroutine. This is the only way to get external data into the UI
	// when Run() owns the event loop.
	go func() {
		for i := 1; ; i++ {
			time.Sleep(200 * time.Millisecond)
			if comp.IsPaused() {
				continue
			}
			msg := fmt.Sprintf("[%s] Message #%d", time.Now().Format("15:04:05.000"), i)
			app.QueueUpdate(func() {
				comp.AddMessage(msg)
			})
		}
	}()

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// stepMode uses Open/Step/Close. You own the frame loop and can do
// arbitrary work between steps. Here the producer's channel is drained
// directly in the loop, so no QueueUpdate is needed.
func stepMode() {
	comp := NewFeedApp("Step()")
	app, err := tui.NewApp(tui.WithRootComponent(comp))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := app.Open(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer app.Close()

	msgCh := startProducer(comp.IsPaused)
	ticker := time.NewTicker(16 * time.Millisecond)
	defer ticker.Stop()

	for {
		// Wait for next frame (acts as a frame rate limiter)
		select {
		case <-ticker.C:
		case <-app.StopCh():
			return
		}

		// Drain pending messages from the producer. This runs on the
		// main goroutine so we can mutate state directly.
	drain:
		for {
			select {
			case msg := <-msgCh:
				comp.AddMessage(msg)
			default:
				break drain
			}
		}

		if !app.Step() {
			return
		}
	}
}

// selectMode uses Open/Events/Dispatch/Render/Close. The producer's channel
// is a first-class case in the select alongside go-tui events. This is the
// most natural way to integrate external event sources.
func selectMode() {
	comp := NewFeedApp("Select()")
	app, err := tui.NewApp(tui.WithRootComponent(comp))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := app.Open(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer app.Close()

	msgCh := startProducer(comp.IsPaused)

	for {
		select {
		case ev := <-app.Events():
			app.Dispatch(ev)
		case msg := <-msgCh:
			comp.AddMessage(msg)
		case <-app.StopCh():
			return
		}
		app.Render()
	}
}
