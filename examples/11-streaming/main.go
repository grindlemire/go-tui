// Package main demonstrates streaming data with channels and timers.
//
// This shows:
// - Struct components with KeyMap-based key handling
// - onChannel={tui.Watch(ch, handler)} for channel watchers
// - onTimer={tui.OnTimer(duration, handler)} for timer watchers
// - Refs for imperative access to elements
//
// To build and run:
//
//	go run ../../cmd/tui generate streaming.gsx
//	go run .
package main

import (
	"fmt"
	"os"
	"time"

	tui "github.com/grindlemire/go-tui"
)

//go:generate go run ../../cmd/tui generate streaming.gsx

func main() {
	dataCh := make(chan string, 100)

	app, err := tui.NewApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create app: %v\n", err)
		os.Exit(1)
	}
	defer app.Close()

	app.SetRootComponent(Streaming(dataCh))

	go produce(dataCh, app.StopCh())

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "App error: %v\n", err)
		os.Exit(1)
	}
}

func produce(ch chan<- string, stopCh <-chan struct{}) {
	defer close(ch)

	messages := []string{
		"Starting up...",
		"Loading configuration...",
		"Connecting to services...",
		"Ready!",
	}

	for _, msg := range messages {
		select {
		case <-stopCh:
			return
		case ch <- fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), msg):
		}
		time.Sleep(300 * time.Millisecond)
	}

	for i := 1; i <= 50; i++ {
		select {
		case <-stopCh:
			return
		case ch <- fmt.Sprintf("[%s] Processing item %d...", time.Now().Format("15:04:05"), i):
		}
		time.Sleep(200 * time.Millisecond)
	}

	select {
	case <-stopCh:
		return
	case ch <- fmt.Sprintf("[%s] Done!", time.Now().Format("15:04:05")):
	}
}
