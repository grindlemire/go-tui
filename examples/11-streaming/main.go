// Package main demonstrates streaming data with channels and auto-scroll.
//
// To build and run:
//
//	go run ../../cmd/tui generate streaming.gsx
//	go run .
package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	tui "github.com/grindlemire/go-tui"
)

//go:generate go run ../../cmd/tui generate streaming.gsx

func main() {
	dataCh := make(chan string, 100)

	app, err := tui.NewApp(
		tui.WithRootComponent(Streaming(dataCh)),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create app: %v\n", err)
		os.Exit(1)
	}
	defer app.Close()

	go produce(dataCh, app.StopCh())

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "App error: %v\n", err)
		os.Exit(1)
	}
}

func produce(ch chan<- string, stopCh <-chan struct{}) {
	defer close(ch)

	metrics := []string{"cpu", "mem", "net", "disk", "io"}

	for {
		select {
		case <-stopCh:
			return
		default:
		}

		metric := metrics[rand.Intn(len(metrics))]
		value := rand.Intn(100)
		ts := time.Now().Format("15:04:05.000")

		var line string
		switch metric {
		case "cpu":
			line = fmt.Sprintf("[%s] cpu:  %d%%", ts, value)
		case "mem":
			line = fmt.Sprintf("[%s] mem:  %.1fG", ts, float64(value)/10.0)
		case "net":
			line = fmt.Sprintf("[%s] net:  %d req/s", ts, value*5)
		case "disk":
			line = fmt.Sprintf("[%s] disk: %d%% used", ts, 40+value/2)
		case "io":
			line = fmt.Sprintf("[%s] io:   %d MB/s", ts, value*2)
		}

		select {
		case <-stopCh:
			return
		case ch <- line:
		}

		time.Sleep(time.Duration(100+rand.Intn(400)) * time.Millisecond)
	}
}
