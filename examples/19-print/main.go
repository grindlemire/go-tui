// Package main demonstrates single-frame printing with go-tui.
//
// This example renders a styled build report inline with normal
// terminal output. No interactive event loop is started.
//
// To build and run:
//
//	go run ../../cmd/tui generate report.gsx
//	go run .
package main

import (
	"fmt"

	tui "github.com/grindlemire/go-tui"
)

//go:generate go run ../../cmd/tui generate report.gsx

func main() {
	fmt.Println("[2026-03-01 10:14:02] starting build...")
	fmt.Println("[2026-03-01 10:14:02] compiling cmd/myapp")
	fmt.Println("[2026-03-01 10:14:03] compiling internal/server")
	fmt.Println("[2026-03-01 10:14:04] running tests...")
	fmt.Println()

	tui.Print(BuildReport("myapp", "PASS", "2.3s", 42, 42))

	fmt.Println()
	fmt.Println("[2026-03-01 10:14:06] uploading artifacts to s3://builds/myapp/latest")
	fmt.Println("[2026-03-01 10:14:07] done.")
}
