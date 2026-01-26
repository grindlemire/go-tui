//go:build ignore

package main

// This file exists solely to provide a go:generate directive at the project root.
// Run `go generate` to regenerate all *_tui.go files from .tui sources.
//
// Usage:
//   go generate
//
// For individual packages, add this directive to any Go file:
//   //go:generate go run github.com/grindlemire/go-tui/cmd/tui generate ./...

//go:generate go run ./cmd/tui generate ./...
