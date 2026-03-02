// Package main demonstrates single-frame printing with go-tui.
//
// This example renders a styled build report once to stdout and exits.
// No interactive event loop is started.
//
// To build and run:
//
//	go run ../../cmd/tui generate report.gsx
//	go run .
package main

import tui "github.com/grindlemire/go-tui"

//go:generate go run ../../cmd/tui generate report.gsx

func main() {
	tui.Print(BuildReport("myapp", "PASS", "2.3s", 42, 42))
}
